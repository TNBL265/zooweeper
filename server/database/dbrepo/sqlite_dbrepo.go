package zooweeper

import (
	"database/sql"
	"log"
	"time"

	"github.com/tnbl265/zooweeper/database/models"
)

type SQLiteDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *SQLiteDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *SQLiteDBRepo) AllMetadata() ([]*models.Metadata, error) {
	rows, err := m.DB.Query("SELECT SenderIp, ReceiverIp, Timestamp, Attempts FROM score")
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil, err
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []*models.Metadata

	for rows.Next() {
		var data models.Metadata
		err := rows.Scan(&data.SenderIp, &data.ReceiverIp, &data.Timestamp, &data.Attempts)
		if err != nil {
			log.Println("Error scanning data", err)
			return nil, err
		}
		results = append(results, &data)
	}

	return results, nil
}

// Updata db via sth query.
func (m *SQLiteDBRepo) InsertMetadata(metadata models.Metadata) error {
	sqlStatement := `
	INSERT INTO score (SenderIp, ReceiverIp, Timestamp, Attempts)
	VALUES (?, ?, ?, ?)
`

	row, err := m.DB.Prepare(sqlStatement)
	if err != nil {
		log.Println("Error preparing row", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(metadata.SenderIp, metadata.ReceiverIp, metadata.Timestamp, metadata.Attempts)
	if err != nil {
		log.Println("Error executing insert row", err)
		return err
	}

	return nil
}
