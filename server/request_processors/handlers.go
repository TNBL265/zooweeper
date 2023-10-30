package zooweeper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tnbl265/zooweeper/database/models"
)

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
		Metadata    models.Metadata    `json:"Metadata"`
		GameResults models.GameResults `json:"GameResults"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	metadata := models.Data{
		Metadata: models.Metadata{
			SenderIp:   requestPayload.Metadata.SenderIp,
			ReceiverIp: requestPayload.Metadata.ReceiverIp,
			Attempts:   requestPayload.Metadata.Attempts,
			Timestamp:  requestPayload.Metadata.Timestamp,
		},
		GameResults: models.GameResults{
			Minute: requestPayload.GameResults.Minute,
			Player: requestPayload.GameResults.Player,
			Club:   requestPayload.GameResults.Club,
			Score:  requestPayload.GameResults.Score,
		},
	}

	// get servers from 'servers' header in table
	ports, _ := app.DB.GetServers()
	ports = []string{"9090"}
	// perform POST request to all servers mentioned.
	jsonData, err := json.Marshal(metadata.GameResults)
	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%s/updateScore", port)
		_ = app.makeExternalRequest(w, url, "POST", jsonData)
	}

	jsonData, err = json.Marshal(metadata.Metadata)
	zkPorts := []int{8080}
	for _, zkPorts := range zkPorts {
		url := fmt.Sprintf("http://localhost:%s/metadata", strconv.Itoa(zkPorts))
		_ = app.makeExternalRequest(w, url, "POST", jsonData)
	}

	app.writeJSON(w, http.StatusOK, metadata)
}

func (app *Application) UpdateMetaData(w http.ResponseWriter, r *http.Request) {
	var requestPayload models.Metadata

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	metadata := models.Metadata{
		SenderIp:   requestPayload.SenderIp,
		ReceiverIp: requestPayload.ReceiverIp,
		Attempts:   requestPayload.Attempts,
		Timestamp:  requestPayload.Timestamp,
	}

	err = app.DB.InsertMetadata(metadata)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	app.writeJSON(w, http.StatusOK, metadata)
}

/*
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
*/

func (app *Application) doesScoreExist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leaderServer")

	result, err := app.DB.CheckMetadataExist(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, strconv.FormatBool(result))
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

}

func (app *Application) DeleteScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leaderServer")

	err := app.DB.DeleteMetadata(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}
