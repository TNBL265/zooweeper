package zooweeper

import (
	"database/sql"
	"log"
	"strings"
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

func (m *SQLiteDBRepo) CheckMetadataExist(leaderServer string) (bool, error) {

	sqlStatement := "SELECT COUNT(*) FROM znode WHERE LeaderServer = ?"

	var count int
	err := m.DB.QueryRow(sqlStatement, leaderServer).Scan(&count)
	if err != nil {
		log.Println("Error executing reading row", err)
		return false, err
	}

	return true, nil
}

func (m *SQLiteDBRepo) DeleteMetadata(leaderServer string) error {

	sqlStatement := "DELETE FROM znode WHERE LeaderServer = ?"

	_, err := m.DB.Exec(sqlStatement, leaderServer)
	if err != nil {
		log.Println("Error executing delete row", err)
		return err
	}

	return nil
}

func (m *SQLiteDBRepo) GetServers() ([]string, error) {
	sqlStatement := "SELECT Servers FROM znode WHERE NodeId = ?"
	rows, err := m.DB.Query(sqlStatement, 1)
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
