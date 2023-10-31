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
