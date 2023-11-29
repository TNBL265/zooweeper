package ztree

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
)

func (zt *ZTree) AllMetadata() ([]*Metadata, error) {
	rows, err := zt.DB.Query("SELECT * FROM ZNode")
	if err != nil {
		log.Println("Error querying the ztree:", err)
		return nil, err
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []*Metadata

	for rows.Next() {
		var data Metadata
		err := rows.Scan(
			&data.NodeId, &data.NodePort, &data.Leader, &data.Servers,
			&data.Timestamp, &data.Version, &data.ParentId,
			&data.Clients, &data.SenderIp, &data.ReceiverIp,
		)
		if err != nil {
			log.Println("Error scanning data", err)
			return nil, err
		}
		results = append(results, &data)
	}

	return results, nil
}

func (zt *ZTree) InsertFirstMetadata(metadata Metadata) error {
	sqlStatement := `
	INSERT INTO ZNode (NodePort, Leader, Servers, Timestamp, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`

	row, err := zt.DB.Prepare(sqlStatement)
	if err != nil {
		log.Println("Error preparing row", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(
		metadata.NodePort, metadata.Leader, metadata.Servers, metadata.Timestamp,
		metadata.Version, 0,
		metadata.Clients, metadata.SenderIp, metadata.ReceiverIp,
	)
	if err != nil {
		log.Println("Error executing insert row", err)
		return err
	}

	return nil
}

// InsertMetadataWithParent
func (zt *ZTree) InsertMetadataWithParent(metadata Metadata) error {
	nodeId, _ := zt.getParentNodeId(metadata.SenderIp)

	if nodeId == 0 {
		// Insert parent process with parentId=1 (direct child of Zookeeper)
		err := zt.insertParentProcessMetadata(metadata)
		if err != nil {
			return err
		}
	} else {
		version, matched, _ := zt.checkSenderClientsMatch(metadata.SenderIp, metadata.Clients)
		if !matched {
			err := zt.updateProcessMetadata(metadata, nodeId, version+1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (zt *ZTree) GetClients(client string) ([]string, error) {
	sqlStatement := "SELECT Clients FROM ZNode WHERE SenderIp=$1"
	rows, err := zt.DB.Query(sqlStatement, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientsStr string
	if rows.Next() {
		err := rows.Scan(&clientsStr)
		if err != nil {
			return nil, err
		}
	}

	clients := strings.Split(clientsStr, ",")

	return clients, nil
}

func (zt *ZTree) InsertMetadata(metadata Metadata) error {
	sqlInsert := `
        INSERT INTO ZNode (NodeId, NodePort, Leader, Servers, Timestamp, Version, ParentId, Clients, SenderIp, ReceiverIp)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := zt.DB.Exec(sqlInsert, metadata.NodeId, metadata.NodePort, metadata.Leader, metadata.Servers, metadata.Timestamp, metadata.Version, metadata.ParentId, metadata.Clients, metadata.SenderIp, metadata.ReceiverIp)
	return err
}

func (zt *ZTree) UpdateFirstLeader(leader string) error {
	leaderNum, err := strconv.Atoi(leader)
	if err != nil {
		return err
	}
	var servers []string
	for i := 8080; i <= leaderNum; i++ {
		servers = append(servers, strconv.Itoa(i))
	}
	serversList := strings.Join(servers, ",")

	sqlStatement := `
		UPDATE ZNode 
		SET Leader = $1, Servers = ?
		WHERE NodeId = "1"
	`
	_, err = zt.DB.Exec(sqlStatement, leader, serversList)
	if err != nil {
		return err
	}
	return err
}

func (zt *ZTree) Connection() *sql.DB {
	return zt.DB
}

func (zt *ZTree) InitializeDB() {
	log.Println("Creating ZNode tables")
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ZNode (
		NodeId INTEGER PRIMARY KEY AUTOINCREMENT,
		NodePort TEXT,
		Leader TEXT,
		Servers TEXT,
		Timestamp DATETIME,
		Version INTEGER,
		ParentId INTEGER,
		Clients TEXT,
		SenderIp TEXT,
		ReceiverIp TEXT
);`

	_, err := zt.DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal("InitializeDB: ", err)
	}
}

func (zt *ZTree) ZNodeIdExists(nodeId int) (bool, error) {
	checkStatement := `SELECT COUNT(*) FROM ZNode WHERE NodeId = ?;`
	var count int
	err := zt.DB.QueryRow(checkStatement, nodeId).Scan(&count)
	if err != nil {
		log.Println("Error checking NodeId existence", err)
		return false, err
	}
	return count > 0, nil
}

func (zt *ZTree) GetLocalMetadata() (*Metadata, error) {
	row := zt.DB.QueryRow("SELECT * FROM ZNode WHERE NodeId = 1")

	var data Metadata
	err := row.Scan(
		&data.NodeId, &data.NodePort, &data.Leader, &data.Servers,
		&data.Timestamp, &data.Version, &data.ParentId,
		&data.Clients, &data.SenderIp, &data.ReceiverIp,
	)
	if err != nil {
		log.Println("Error scanning data:", err)
		return nil, err
	}

	return &data, nil
}

func (zt *ZTree) getParentNodeId(senderIp string) (int, error) {
	sqlCheck := `SELECT NodeId FROM ZNode WHERE SenderIp = ?`
	var nodeId int
	err := zt.DB.QueryRow(sqlCheck, senderIp).Scan(&nodeId)
	if err != nil {
		// does not exist
		return 0, err
	}
	return nodeId, nil
}

func (zt *ZTree) insertParentProcessMetadata(metadata Metadata) error {
	sqlPartialInsert := `
	INSERT INTO ZNode (NodePort, Leader, Servers, Timestamp, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	row, err := zt.DB.Prepare(sqlPartialInsert)
	if err != nil {
		log.Println("Error prepare row for insertParentProcessMetadata:", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(
		"", "", "", metadata.Timestamp, 0, 1,
		metadata.Clients, metadata.SenderIp, metadata.ReceiverIp,
	)
	if err != nil {
		log.Println("Error exec for insertParentProcessMetadata:", err)
		return err
	}

	return nil
}

func (zt *ZTree) checkSenderClientsMatch(senderIp, clients string) (int, bool, error) {
	// First, find the highest NodeId for the given senderIp
	sqlGetHighestNodeId := `
        SELECT NodeId, Version 
        FROM ZNode 
        WHERE SenderIp = ?
        ORDER BY NodeId DESC
        LIMIT 1
    `
	var highestNodeId int
	var version int

	err := zt.DB.QueryRow(sqlGetHighestNodeId, senderIp).Scan(&highestNodeId, &version)
	if err != nil {
		log.Println("Error finding highest NodeId for checkSenderClientsMatch:", err)
		return 0, false, err
	}

	if highestNodeId == 0 {
		return 0, false, nil
	}

	// Second, check if the highest NodeId also matches the given clients
	sqlCheckClients := `
        SELECT 1
        FROM ZNode 
        WHERE NodeId = ? AND Clients = ?
    `
	var exists int
	err = zt.DB.QueryRow(sqlCheckClients, highestNodeId, clients).Scan(&exists)
	if err != nil {
		return version, false, err
	}

	return version, true, nil

}

func (zt *ZTree) updateProcessMetadata(metadata Metadata, parent, version int) error {
	sqlPartialInsert := `
	INSERT INTO ZNode (NodePort, Leader, Servers, Timestamp, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	row, err := zt.DB.Prepare(sqlPartialInsert)
	if err != nil {
		log.Println("Error preparing row for updateClients:", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(
		"", "", "", metadata.Timestamp, 0, version, parent,
		metadata.Clients, metadata.SenderIp, metadata.ReceiverIp,
	)
	if err != nil {
		log.Println("Error exec for updateClients:", err)
		return err
	}

	return nil
}

func (zt *ZTree) GetHighestZNodeId() (int, error) {
	sqlStatement := `
        SELECT MAX(NodeId) FROM ZNode
    `
	var highestZNodeId int
	err := zt.DB.QueryRow(sqlStatement).Scan(&highestZNodeId)
	if err != nil {
		log.Println("Error retrieving the highest NodeId:", err)
		return 0, err
	}
	return highestZNodeId, nil
}

func (zt *ZTree) GetMetadatasGreaterThanZNodeId(highestZNodeId int) (Metadatas, error) {
	sqlStatement := `
        SELECT NodeId, NodePort, Leader, Servers, Timestamp, Version, ParentId, Clients, SenderIp, ReceiverIp
        FROM ZNode
        WHERE NodeId > ?
    `
	rows, err := zt.DB.Query(sqlStatement, highestZNodeId)
	if err != nil {
		log.Println("Error querying Metadatas:", err)
		return Metadatas{}, err
	}
	defer rows.Close()

	var metadatas Metadatas
	for rows.Next() {
		var md Metadata
		err := rows.Scan(&md.NodeId, &md.NodePort, &md.Leader, &md.Servers, &md.Timestamp, &md.Version, &md.ParentId, &md.Clients, &md.SenderIp, &md.ReceiverIp)
		if err != nil {
			log.Println("Error scanning Metadata row:", err)
			return Metadatas{}, err
		}
		metadatas.MetadataList = append(metadatas.MetadataList, md)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating through Metadata rows:", err)
		return Metadatas{}, err
	}

	return metadatas, nil
}
