package zooweeper

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type WriteOps struct {
	ab *AtomicBroadcast
}

func (wo *WriteOps) UpdateMetaData(http.ResponseWriter, *http.Request) {
}

func (wo *WriteOps) WriteMetaData(w http.ResponseWriter, r *http.Request) {
	data := wo.ab.CreateMetadata(w, r)
	err := wo.ab.ZTree.UpsertMetadata(data.Metadata)
	if err != nil {
		wo.ab.errorJSON(w, err, http.StatusBadRequest)
		log.Fatal(err)
		return
	}

	// Only modify Kafka-Server metadata if it is a leader
	zNode, _ := wo.ab.ZTree.GetLocalMetadata()
	if zNode.NodeIp == zNode.Leader {
		ports, _ := wo.ab.ZTree.GetClients()
		jsonData, _ := json.Marshal(data.GameResults)
		for _, port := range ports {
			url := fmt.Sprintf("%s:%s/updateScore", wo.ab.BaseURL, port)
			_, _ = wo.ab.makeExternalRequest(w, url, "POST", jsonData)
		}
	}

	wo.ab.writeJSON(w, http.StatusOK, data)
}

func (wo *WriteOps) DeleteScore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "leader")

	err := wo.ab.ZTree.DeleteMetadata(id)
	if err != nil {
		wo.ab.errorJSON(w, err)
		return
	}
}
