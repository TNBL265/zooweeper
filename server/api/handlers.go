package zooweeper

import (
	"fmt"
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
	results, err := app.DB.AllGameResults()

	// return results
	err = app.writeJSON(w, http.StatusOK, results)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}
