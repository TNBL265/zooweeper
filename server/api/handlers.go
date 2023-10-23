package zooweeper

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

type GameResults struct {
	Minute int    `json:"Minute"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

func (app *Application) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func (app *Application) GetGameResults(w http.ResponseWriter, r *http.Request) {
	// connect to the database.
	db, err := sql.Open("sqlite3", "zooweeper-database.db")
	if err != nil {
		http.Error(w, "Unable to connect to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// select all from table
	rows, err := db.Query("SELECT Minute, Player, Club, Score FROM game_results")
	if err != nil {
		log.Println("Error querying the database:", err)
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// collate all rows into one slice
	var results []GameResults

	for rows.Next() {
		var data GameResults
		err := rows.Scan(&data.Minute, &data.Player, &data.Club, &data.Score)
		if err != nil {
			http.Error(w, "Error scanning data", http.StatusInternalServerError)
			return
		}
		results = append(results, data)
	}

	// return results
	err = app.writeJSON(w, http.StatusOK, results)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}
