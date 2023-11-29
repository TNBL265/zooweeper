package zooweeper

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/tnbl265/zooweeper/ztree"
	"log"
	"net/http"
	"strings"
	"time"
)

type SyncOps struct {
	ab *AtomicBroadcast
}

// SyncRequestHandler send back current highest ZNodeId
func (so *SyncOps) SyncRequestHandler(_ http.ResponseWriter, r *http.Request) {
	zNode, _ := so.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")

	highestZNodeId, _ := so.ab.ZTree.GetHighestZNodeId()
	var metadata = &ztree.Metadata{
		NodeId: highestZNodeId,
	}
	jsonData, _ := json.Marshal(metadata)

	color.Yellow("%s received SyncRequest from %s", zNode.NodeIp, clientPort)
	color.Yellow("%s sending syncACK with highest ZNodeId %d to %s\n", zNode.NodeIp, highestZNodeId, clientPort)
	url := fmt.Sprintf(so.ab.BaseURL + ":" + clientPort + "/syncResponse")
	_, err = so.ab.makeExternalRequest(nil, url, "POST", jsonData)
}

// SyncResponseHandler wait for majority value of highest ZNodeId and send back current highestZNodeId
func (so *SyncOps) SyncResponseHandler(w http.ResponseWriter, r *http.Request) {
	zNode, _ := so.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")

	var requestPayload ztree.Metadata
	so.ab.readJSON(w, r, &requestPayload)

	color.Yellow("%s received SyncResponse with highestZNodeId %d from %s", zNode.NodeIp, requestPayload.NodeId, clientPort)

	// Wait for majority of Follower to ACK
	portsSlice := strings.Split(zNode.Servers, ",")
	majority := len(portsSlice) / 2

	if so.ab.SyncState() != ACKED {
		for {
			currentSyncAckCount := so.ab.SyncCounter()
			currentSyncAckCount++
			so.ab.SetSyncCounter(currentSyncAckCount)

			if currentSyncAckCount > majority {
				color.Yellow("Leader %s received majority syncAck, %d\n", zNode.NodeIp, currentSyncAckCount)
				so.ab.SetSyncCounter(0)
				so.ab.SetSyncState(ACKED)
				break
			} else if so.ab.SyncState() == ACKED {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// RequestMetadataHandler send requested Metadatas to Leader to update
func (so *SyncOps) RequestMetadataHandler(w http.ResponseWriter, r *http.Request) {
	zNode, _ := so.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")

	var requestPayload ztree.Metadata
	so.ab.readJSON(w, r, &requestPayload)

	highestZNodeId := requestPayload.NodeId
	color.Yellow("%s received RequestMetadata with highestZNodeId %d from %s", zNode.NodeIp, highestZNodeId, clientPort)

	metadatas, _ := so.ab.ZTree.GetMetadatasGreaterThanZNodeId(highestZNodeId)
	jsonData, _ := json.Marshal(metadatas)

	color.Yellow("%s send requested Metadata to %s\n", zNode.NodeIp, clientPort)
	url := so.ab.BaseURL + ":" + clientPort + "/updateMetadata"
	so.ab.makeExternalRequest(nil, url, "POST", jsonData)
}

// UpdateMetadataHandler for Leader to update Metadata
func (so *SyncOps) UpdateMetadataHandler(w http.ResponseWriter, r *http.Request) {
	zNode, _ := so.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")

	var metadatas ztree.Metadatas
	so.ab.readJSON(w, r, &metadatas)

	color.Yellow("%s received updated Metadata from %s", zNode.NodeIp, clientPort)

	for _, metadata := range metadatas.MetadataList {
		exists, err := so.ab.ZTree.ZNodeIdExists(metadata.NodeId)
		if err != nil {
			log.Println("Error checking NodeId existence:", err)
			continue
		}
		if !exists {
			err := so.ab.ZTree.InsertMetadata(metadata)
			if err != nil {
				log.Println("Error inserting Metadata:", err)
				continue
			}
			color.Yellow("Inserted Metadata for NodeId %d", metadata.NodeId)
		}
	}
	_ = so.ab.writeJSON(w, http.StatusOK, "Updated Metadata")
}
