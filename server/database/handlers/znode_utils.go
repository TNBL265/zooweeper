package handlers

import (
	"database/sql"
	"github.com/tnbl265/zooweeper/database/models"
	"log"
)

func (zt *ZTree) Connection() *sql.DB {
	return zt.DB
}

func (zt *ZTree) InitializeDB() {
	log.Println("Creating ZNode tables")
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ZNode (
		NodeId INTEGER PRIMARY KEY AUTOINCREMENT,
		NodeIp TEXT,
		Leader TEXT,
		Servers TEXT,
		Timestamp DATETIME,
		Attempts INTEGER,
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

func (zt *ZTree) GetLocalMetadata() (*models.Metadata, error) {
	row := zt.DB.QueryRow("SELECT * FROM ZNode WHERE NodeId = 1")

	var data models.Metadata
	err := row.Scan(
		&data.NodeId, &data.NodeIp, &data.Leader, &data.Servers,
		&data.Timestamp, &data.Attempts, &data.Version, &data.ParentId,
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

func (zt *ZTree) insertParentProcessMetadata(metadata models.Metadata) error {
	sqlPartialInsert := `
	INSERT INTO ZNode (NodeIp, Leader, Servers, Timestamp, Attempts, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	row, err := zt.DB.Prepare(sqlPartialInsert)
	if err != nil {
		log.Println("Error prepare row for insertParentProcessMetadata:", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(
		"", "", "", metadata.Timestamp, 0, 0, 1,
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

func (zt *ZTree) updateProcessMetadata(metadata models.Metadata, parent, version int) error {
	sqlPartialInsert := `
	INSERT INTO ZNode (NodeIp, Leader, Servers, Timestamp, Attempts, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
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

func (zt *ZTree) GetMetadatasGreaterThanZNodeId(highestZNodeId int) (models.Metadatas, error) {
	sqlStatement := `
        SELECT NodeId, NodeIp, Leader, Servers, Timestamp, Attempts, Version, ParentId, Clients, SenderIp, ReceiverIp
        FROM ZNode
        WHERE NodeId > ?
    `
	rows, err := zt.DB.Query(sqlStatement, highestZNodeId)
	if err != nil {
		log.Println("Error querying Metadatas:", err)
		return models.Metadatas{}, err
	}
	defer rows.Close()

	var metadatas models.Metadatas
	for rows.Next() {
		var md models.Metadata
		err := rows.Scan(&md.NodeId, &md.NodeIp, &md.Leader, &md.Servers, &md.Timestamp, &md.Attempts, &md.Version, &md.ParentId, &md.Clients, &md.SenderIp, &md.ReceiverIp)
		if err != nil {
			log.Println("Error scanning Metadata row:", err)
			return models.Metadatas{}, err
		}
		metadatas.MetadataList = append(metadatas.MetadataList, md)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating through Metadata rows:", err)
		return models.Metadatas{}, err
	}

	return metadatas, nil
}
