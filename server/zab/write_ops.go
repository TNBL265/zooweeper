package zab

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteOps for WriteRequest
type WriteOps struct {
	ab *AtomicBroadcast
}

// UpdateMetadata is a placeholder as main processing is in WriteOpsMiddleware
func (wo *WriteOps) UpdateMetadata(http.ResponseWriter, *http.Request) {
}

// WriteMetadata handler to write into ZTree using InsertMetadataWithParent
func (wo *WriteOps) WriteMetadata(w http.ResponseWriter, r *http.Request) {
	data := wo.ab.CreateMetadataFromPayload(w, r)
	wo.ab.ZTree.InsertMetadataWithParent(data.Metadata)

	// Only modify Kafka broker metadata if it is a leader
	zNode, _ := wo.ab.ZTree.GetLocalMetadata()
	if zNode.NodePort == zNode.Leader {
		ports, _ := wo.ab.ZTree.GetClients(data.Metadata.SenderIp)
		jsonData, _ := json.Marshal(data.GameResults)
		for _, port := range ports {
			url := fmt.Sprintf("%s:%s/updateScore", wo.ab.BaseURL, port)
			_, _ = wo.ab.sendRequest(url, "POST", jsonData)
		}
	}

	wo.ab.writeJSON(w, http.StatusOK, data)
	wo.ab.SetProposalState(COMMITTED)
}
