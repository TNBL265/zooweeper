package handlers

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/tnbl265/zooweeper/database/models"
)

type ZTree struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (zt *ZTree) Connection() *sql.DB {
	return zt.DB
}

func (zt *ZTree) AllMetadata() ([]*models.Metadata, error) {
	rows, err := zt.DB.Query("SELECT LeaderServer, Servers,  SenderIp, ReceiverIp, Timestamp, Attempts FROM ZNode")
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil, err
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []*models.Metadata

	for rows.Next() {
		var data models.Metadata
		err := rows.Scan(&data.LeaderServer, &data.Servers, &data.SenderIp, &data.ReceiverIp, &data.Timestamp, &data.Attempts)
		if err != nil {
			log.Println("Error scanning data", err)
			return nil, err
		}
		results = append(results, &data)
	}

	return results, nil
}

func (zt *ZTree) InsertMetadata(metadata models.Metadata) error {
	sqlStatement := `
	INSERT INTO ZNode (LeaderServer, Servers, NodeIp, SenderIp, ReceiverIp, Timestamp, Version, Attempts)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`

	row, err := zt.DB.Prepare(sqlStatement)
	if err != nil {
		log.Println("Error preparing row", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec("-", "-", "8080", metadata.SenderIp, metadata.ReceiverIp, metadata.Timestamp, 0, metadata.Attempts)
	if err != nil {
		log.Println("Error executing insert row", err)
		return err
	}

	return nil
}

func (zt *ZTree) CheckMetadataExist(leaderServer string) (bool, error) {

	sqlStatement := "SELECT COUNT(*) FROM ZNode WHERE LeaderServer = ?"

	var count int
	err := zt.DB.QueryRow(sqlStatement, leaderServer).Scan(&count)
	if err != nil {
		log.Println("Error executing reading row", err)
		return false, err
	}

	return true, nil
}

func (zt *ZTree) DeleteMetadata(leaderServer string) error {

	sqlStatement := "DELETE FROM ZNode WHERE LeaderServer = ?"

	_, err := zt.DB.Exec(sqlStatement, leaderServer)
	if err != nil {
		log.Println("Error executing delete row", err)
		return err
	}

	return nil
}

func (zt *ZTree) GetServers() ([]string, error) {
	sqlStatement := "SELECT Servers FROM ZNode WHERE NodeId = ?"
	rows, err := zt.DB.Query(sqlStatement, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serversStr string
	if rows.Next() {
		err := rows.Scan(&serversStr)
		if err != nil {
			return nil, err
		}
	}

	servers := strings.Split(serversStr, ", ")

	return servers, nil
}
