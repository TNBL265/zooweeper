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

func (m *SQLiteDBRepo) AllMetadata() ([]*models.Sello, error) {
	rows, err := m.DB.Query("SELECT LeaderServer, Servers,  SenderIp, ReceiverIp, Timestamp, Attempts FROM znode")
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil, err
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []*models.Sello

	for rows.Next() {
		var data models.Sello
		err := rows.Scan(&data.LeaderServer, &data.Servers, &data.SenderIp, &data.ReceiverIp, &data.Timestamp, &data.Attempts)
		if err != nil {
			log.Println("Error scanning data", err)
			return nil, err
		}
		results = append(results, &data)
	}

	return results, nil
}

// Updata db via sth query.
func (m *SQLiteDBRepo) InsertMetadata(metadata models.Sello) error {
	sqlStatement := `
	INSERT INTO znode (LeaderServer, Servers, SenderIp, ReceiverIp, Timestamp, Attempts)
	VALUES (?, ?, ?, ?, ?, ?)
`

	row, err := m.DB.Prepare(sqlStatement)
	if err != nil {
		log.Println("Error preparing row", err)
		return err
	}
	defer row.Close()

	_, err = row.Exec(metadata.LeaderServer, metadata.Servers, metadata.SenderIp, metadata.ReceiverIp, metadata.Timestamp, metadata.Attempts)
	if err != nil {
		log.Println("Error executing insert row", err)
		return err
	}

	return nil
}
