package zooweeper

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/tnbl265/zooweeper/database/models"
	"net/http"
	"strconv"
)

type WriteOps struct {
	ab *AtomicBroadcast
}

func (wo *WriteOps) AddScore(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Metadata    models.Metadata    `json:"Metadata"`
		GameResults models.GameResults `json:"GameResults"`
	}

	err := wo.ab.readJSON(w, r, &requestPayload)
	if err != nil {
		wo.ab.errorJSON(w, err, http.StatusBadRequest)
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
	ports, _ := wo.ab.ZTree.GetServers()
	ports = []string{"9090"}
	// perform POST request to all servers mentioned.
	jsonData, err := json.Marshal(metadata.GameResults)
	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%s/updateScore", port)
		_ = wo.ab.makeExternalRequest(w, url, "POST", jsonData)
	}

	jsonData, err = json.Marshal(metadata.Metadata)
	zkPorts := []int{8080}
	for _, zkPorts := range zkPorts {
		url := fmt.Sprintf("http://localhost:%s/metadata", strconv.Itoa(zkPorts))
		_ = wo.ab.makeExternalRequest(w, url, "POST", jsonData)
	}

	wo.ab.writeJSON(w, http.StatusOK, metadata)
}

func (wo *WriteOps) UpdateMetaData(w http.ResponseWriter, r *http.Request) {
	var requestPayload models.Metadata

	err := wo.ab.readJSON(w, r, &requestPayload)
	if err != nil {
		wo.ab.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	metadata := models.Metadata{
		SenderIp:   requestPayload.SenderIp,
		ReceiverIp: requestPayload.ReceiverIp,
		Attempts:   requestPayload.Attempts,
		Timestamp:  requestPayload.Timestamp,
	}

	err = wo.ab.ZTree.InsertMetadata(metadata)
	if err != nil {
		wo.ab.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	wo.ab.writeJSON(w, http.StatusOK, metadata)
}

func (wo *WriteOps) DeleteScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leaderServer")

	err := wo.ab.ZTree.DeleteMetadata(id)
	if err != nil {
		wo.ab.errorJSON(w, err)
		return
	}
}
