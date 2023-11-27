package zooweeper

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/tnbl265/zooweeper/database/models"
	"net/http"
)

type SyncOps struct {
	ab *AtomicBroadcast
}

// SyncRequestHandler send back Metadatas to Leader
func (so *SyncOps) SyncRequestHandler(_ http.ResponseWriter, r *http.Request) {
	zNode, _ := so.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")
	if clientPort != zNode.Leader {
		color.Red("I only supposed to receive Sync Request from leader not from %s\n", clientPort)
	}
	color.Yellow("Received SYNC Request from %s", clientPort)

	url := fmt.Sprintf(so.ab.BaseURL + ":" + clientPort + "/updateMetadata")

	metadataList, _ := so.ab.ZTree.GetMetadataWithParentId(1)
	jsonData, _ := json.Marshal(metadataList)

	_, err = so.ab.makeExternalRequest(nil, url, "POST", jsonData)
}

// SyncMetadataHandler for Leader to update Metadata
func (so *SyncOps) SyncMetadataHandler(w http.ResponseWriter, r *http.Request) {
	var metadatas models.Metadatas
	err := so.ab.readJSON(w, r, &metadatas)
	if err != nil {
		so.ab.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	for _, metadata := range metadatas.MetadataList {
		currentVersion, _ := so.ab.ZTree.GetVersionBySenderIp(metadata.SenderIp)
		if metadata.Version > currentVersion {
			color.Yellow("Updating %s to version %d", metadata.SenderIp, metadata.Version)
			color.Yellow("%v", metadata)
			err := so.ab.ZTree.UpdateMetadata(metadata)
			if err != nil {
				color.Red("Error updating metadata: %s", err)
				continue
			}
		}
	}

	_ = so.ab.writeJSON(w, http.StatusOK, "Updated Metadata")

}
