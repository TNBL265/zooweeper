package zooweeper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	ztree "github.com/tnbl265/zooweeper/database"
	"github.com/tnbl265/zooweeper/database/models"
	"net/http"
	"strconv"
)

type AtomicBroadcast struct {
	ZTree ztree.ZooWeeperDatabaseRepo
}

func (ab *AtomicBroadcast) OpenDB(datasource string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", datasource)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (ab *AtomicBroadcast) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func (ab *AtomicBroadcast) GetAllMetadata(w http.ResponseWriter, r *http.Request) {
	// connect to the database.
	results, err := ab.ZTree.AllMetadata()

	// return results
	err = ab.writeJSON(w, http.StatusOK, results)
	if err != nil {
		ab.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
}

func (ab *AtomicBroadcast) AddScore(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Metadata    models.Metadata    `json:"Metadata"`
		GameResults models.GameResults `json:"GameResults"`
	}

	err := ab.readJSON(w, r, &requestPayload)
	if err != nil {
		ab.errorJSON(w, err, http.StatusBadRequest)
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
	ports, _ := ab.ZTree.GetServers()
	ports = []string{"9090"}
	// perform POST request to all servers mentioned.
	jsonData, err := json.Marshal(metadata.GameResults)
	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%s/updateScore", port)
		_ = ab.makeExternalRequest(w, url, "POST", jsonData)
	}

	jsonData, err = json.Marshal(metadata.Metadata)
	zkPorts := []int{8080}
	for _, zkPorts := range zkPorts {
		url := fmt.Sprintf("http://localhost:%s/metadata", strconv.Itoa(zkPorts))
		_ = ab.makeExternalRequest(w, url, "POST", jsonData)
	}

	ab.writeJSON(w, http.StatusOK, metadata)
}

func (ab *AtomicBroadcast) UpdateMetaData(w http.ResponseWriter, r *http.Request) {
	var requestPayload models.Metadata

	err := ab.readJSON(w, r, &requestPayload)
	if err != nil {
		ab.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	metadata := models.Metadata{
		SenderIp:   requestPayload.SenderIp,
		ReceiverIp: requestPayload.ReceiverIp,
		Attempts:   requestPayload.Attempts,
		Timestamp:  requestPayload.Timestamp,
	}

	err = ab.ZTree.InsertMetadata(metadata)
	if err != nil {
		ab.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	ab.writeJSON(w, http.StatusOK, metadata)
}

func (ab *AtomicBroadcast) DoesScoreExist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leaderServer")

	result, err := ab.ZTree.CheckMetadataExist(id)
	if err != nil {
		ab.errorJSON(w, err)
		return
	}

	err = ab.writeJSON(w, http.StatusOK, strconv.FormatBool(result))
	if err != nil {
		ab.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

}

func (ab *AtomicBroadcast) DeleteScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leaderServer")

	err := ab.ZTree.DeleteMetadata(id)
	if err != nil {
		ab.errorJSON(w, err)
		return
	}
}
