package zooweeper

import (
	"database/sql"
	"log"
	"time"

	"github.com/tnbl265/zooweeper/server/database/models"
)

type SQLiteDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *SQLiteDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *SQLiteDBRepo) AllGameResults() ([]*models.GameResults, error) {
	rows, err := m.DB.Query("SELECT Minute, Player, Club, Score FROM game_results")
	if err != nil {
		log.Println("Error querying the database:", err)
		return nil, err
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []*models.GameResults

	for rows.Next() {
		var data models.GameResults
		err := rows.Scan(&data.Minute, &data.Player, &data.Club, &data.Score)
		if err != nil {
			log.Println("Error scanning data", err)
			return nil, err
		}
		results = append(results, &data)
	}

	return results, nil
}
