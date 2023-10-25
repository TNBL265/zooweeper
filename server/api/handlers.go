package zooweeper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/tnbl265/zooweeper/database/models"
)

const frontendUrl = "http://localhost:9092"

func (app *Application) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func (app *Application) GetAllMetadata(w http.ResponseWriter, r *http.Request) {
	// connect to the database.
	results, err := app.DB.AllMetadata()

	// return results
	err = app.writeJSON(w, http.StatusOK, results)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}

func (app *Application) AddScore(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		LeaderServer string             `json:"LeaderServer"`
		Servers      string             `json:"Servers"`
		SenderIp     string             `json:"SenderIp"`
		ReceiverIp   string             `json:"ReceiverIp"`
		Attempts     int                `json:"Attempts"`
		Event        models.GameResults `json:"Event"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	timeReceived := time.Now()

	// insert metadata into database
	metadata := models.Sello{
		ServersData: models.ServersData{
			LeaderServer: requestPayload.LeaderServer,
			Servers:      requestPayload.Servers,
		},
		Metadata: models.Metadata{
			SenderIp:   requestPayload.SenderIp,
			ReceiverIp: requestPayload.ReceiverIp,
			Attempts:   requestPayload.Attempts,
			Timestamp:  timeReceived,
		},
	}
	err = app.DB.InsertMetadata(metadata)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// post event to nodejs server.
	jsonData, err := json.Marshal(requestPayload.Event)
	url := fmt.Sprintf("%s/updateScore", frontendUrl)

	res := app.makeExternalRequest(w, url, "POST", jsonData)
	defer res.Body.Close()

	// return response from frontend,
	// TODO:to remove?
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading request's response:", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// NOTE: take note of receiving type expected, change as needed.
	var responseObject models.GameResults
	if err := json.Unmarshal(bodyBytes, &responseObject); err != nil {
		log.Println("Error unmarshaling JSON response:", err)
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.writeJSON(w, http.StatusOK, responseObject)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}
