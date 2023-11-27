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
	exists, err := zt.parentProcessExist(metadata.SenderIp)
	if err != nil {
		log.Println("Error checking entry existence:", err)
		return err
	}

	if !exists {
		// Insert parent process with parentId=1 (direct child of Zookeeper)
		err = zt.insertParentProcessMetadata(metadata)
		if err != nil {
			return err
		}
	}

	// Insert actual metadata with ParentId=2 (child of above parent process)
	err = zt.InsertMetadata(metadata, 2)
	if err != nil {
		return err
	}

	return nil
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

func (zt *ZTree) GetClients() ([]string, error) {
	sqlStatement := "SELECT Clients FROM ZNode WHERE NodeId = ?"
	rows, err := zt.DB.Query(sqlStatement, 2)
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

	servers := strings.Split(serversStr, ",")

	return servers, nil
}

func (zt *ZTree) UpdateFirstLeader(leader string) error {
	sqlStatement := `
		UPDATE ZNode 
		SET Leader = $1 
	`
	result, err := zt.DB.Exec(sqlStatement, leader)
	if err != nil {
		log.Println("Error updating Leader column:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return err
	}

	if rowsAffected == 0 {
		log.Println("No rows were updated. The table might be empty.")
	}

	return nil
}
