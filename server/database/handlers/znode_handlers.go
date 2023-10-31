package handlers

import (
	"database/sql"
	"github.com/tnbl265/zooweeper/database/models"
	"log"
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

func (zt *ZTree) InsertMetadata(metadata models.Metadata) error {
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
		metadata.Attempts, metadata.Version, metadata.ParentId,
		metadata.Clients, metadata.SenderIp, metadata.ReceiverIp,
	)
	if err != nil {
		log.Println("Error executing insert row", err)
		return err
	}

	return nil
}

func (zt *ZTree) CheckMetadataExist(leader string) (bool, error) {

	sqlStatement := "SELECT COUNT(*) FROM ZNode WHERE Leader = ?"

	var count int
	err := zt.DB.QueryRow(sqlStatement, leader).Scan(&count)
	if err != nil {
		log.Println("Error executing reading row", err)
		return false, err
	}

	return true, nil
}

func (zt *ZTree) DeleteMetadata(leader string) error {

	sqlStatement := "DELETE FROM ZNode WHERE Leader = ?"

	_, err := zt.DB.Exec(sqlStatement, leader)
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
