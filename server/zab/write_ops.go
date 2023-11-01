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
	err := wo.ab.ZTree.InsertMetadata(data.Metadata)
	if err != nil {
		wo.ab.errorJSON(w, err, http.StatusBadRequest)
		log.Fatal(err)
		return
	}

	// Only modify Kafka-Server metadata if it is a leader
	zNode, _ := wo.ab.ZTree.GetLocalMetadata()
	if zNode.NodeIp == zNode.Leader {
		ports, _ := wo.ab.ZTree.GetServers()
		ports = []string{"9090"}
		jsonData, _ := json.Marshal(data.GameResults)
		for _, port := range ports {
			log.Println(port)
			url := fmt.Sprintf("http://localhost:%s/updateScore", port)
			_ = wo.ab.makeExternalRequest(w, url, "POST", jsonData)
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
