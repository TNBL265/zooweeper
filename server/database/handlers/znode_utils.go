package handlers

import (
	"database/sql"
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
