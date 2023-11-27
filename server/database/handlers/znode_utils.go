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

func (zt *ZTree) NodeIdExists(nodeId int) (bool, error) {
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

func (zt *ZTree) parentProcessExist(senderIp string) (bool, error) {
	sqlCheck := `SELECT COUNT(*) FROM ZNode WHERE SenderIp = ?`
	var count int
	err := zt.DB.QueryRow(sqlCheck, senderIp).Scan(&count)
	if err != nil {
		log.Println("Error checking row existence:", err)
		return false, err
	}
	return count > 0, nil
}

func (zt *ZTree) insertParentProcessMetadata(metadata models.Metadata) error {
	sqlPartialInsert := `
	INSERT INTO ZNode (NodeIp, Leader, Servers, Timestamp, Attempts, Version, ParentId, Clients, SenderIp, ReceiverIp) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	row, err := zt.DB.Prepare(sqlPartialInsert)
	if err != nil {
		log.Println("Error preparing row for partial insert:", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(
		metadata.NodeIp, "", "", "", 0, 0, 1,
		metadata.Clients, metadata.SenderIp, metadata.ReceiverIp,
	)
	if err != nil {
		log.Println("Error executing partial insert:", err)
		return err
	}

	return nil
}

func (zt *ZTree) checkSenderClientsMatch(senderIp, clients string) (bool, error) {
	sqlCheck := `
        SELECT COUNT(*) 
        FROM ZNode 
        WHERE SenderIp = ? AND Clients = ?
    `
	var count int
	err := zt.DB.QueryRow(sqlCheck, senderIp, clients).Scan(&count)
	if err != nil {
		log.Println("Error checking for matching SenderIp and Clients:", err)
		return false, err
	}
	return count > 0, nil
}

func (zt *ZTree) updateClients(senderIp, clients string) error {
	sqlStatement := `
        UPDATE ZNode 
        SET Clients = $1, Version = Version + 1
        where SenderIp = $2 AND parentId = 1
    `
	result, err := zt.DB.Exec(sqlStatement, clients, senderIp)
	if err != nil {
		log.Println("Error updating Clients and Version columns:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return err
	}

	if rowsAffected == 0 {
		log.Println("No rows were updated. The table might be empty or the clients are the same as before.")
	}

	return nil
}
