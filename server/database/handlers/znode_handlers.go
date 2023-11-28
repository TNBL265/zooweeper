package handlers

import (
	"database/sql"
	"github.com/tnbl265/zooweeper/database/models"
	"log"
	"strconv"
	"strings"
)

type ZTree struct {
	DB *sql.DB
}

func (zt *ZTree) AllMetadata() ([]*models.Metadata, error) {
	rows, err := zt.DB.Query("SELECT * FROM ZNode")
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil, err
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []*models.Metadata

	for rows.Next() {
		var data models.Metadata
		err := rows.Scan(
			&data.NodeId, &data.NodeIp, &data.Leader, &data.Servers,
			&data.Timestamp, &data.Attempts, &data.Version, &data.ParentId,
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

func (zt *ZTree) InsertMetadata(metadata models.Metadata, parentId int) error {
	sqlStatement := `
	INSERT INTO ZNode (NodeIp, Leader, Servers, Timestamp, Attempts, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`

	row, err := zt.DB.Prepare(sqlStatement)
	if err != nil {
		log.Println("Error preparing row", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(
		metadata.NodeIp, metadata.Leader, metadata.Servers, metadata.Timestamp,
		metadata.Attempts, metadata.Version, parentId,
		metadata.Clients, metadata.SenderIp, metadata.ReceiverIp,
	)
	if err != nil {
		log.Println("Error executing insert row", err)
		return err
	}

	return nil
}

func (zt *ZTree) UpsertMetadata(metadata models.Metadata) error {
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
