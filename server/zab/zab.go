// package zab implements the Atomic Broadcast component for our ZooWeeper.
//
// 1. All Write Request are forwarded to the Leader while Read Request are done locally (details in WriteOpsMiddleware)
// 2. Make use of simple majority quorum to decide on proposal for Data Synchronization
// 3. Instead of using TCP, we use HTTP with the additional of a QueueMiddleware to ensure FIFO client order by ordering
// Request by received Timestamp
// 4. Each transaction from client (Kafka-Server) would be recorded into ZTree as a ZNode

package zab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/tnbl265/zooweeper/request_processors/data"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tnbl265/zooweeper/ztree"
)

// AtomicBroadcast main components of ZooWeeper, defining all request handlers and operations
type AtomicBroadcast struct {
	BaseURL string

	Read     ReadOps
	Write    WriteOps
	Proposal ProposalOps
	Election ElectionOps
	Sync     SyncOps

	ZTree ztree.ZNodeHandlers

	// Proposal
	ackCounter    int
	proposalState ProposalState
	proposalMu    sync.Mutex

	// DataSync
	syncCounter int
	syncState   SyncState
	syncMu      sync.Mutex

	pq PriorityQueue

	ErrorLeaderChan chan data.HealthCheckError
}

// ProposalState for 2PC of Write Request (Ref: Active Messaging in https://zookeeper.apache.org/doc/current/zookeeperInternals.html)
type ProposalState string

const (
	PROPOSED     ProposalState = "PROPOSED"
	ACKNOWLEDGED ProposalState = "ACKNOWLEDGED"
	COMMITTED    ProposalState = "COMMITTED"
)

// SyncState using same logic as ProposalState
type SyncState string

const (
	PREPARED SyncState = "PREPARED"
	ACKED    SyncState = "ACKED"
	SYNCED   SyncState = "SYNCED"
)

var err error

// StartHealthCheck by pinging all other servers to perform HealthCheck, write data to ErrorLeaderChan upon timing out
func (ab *AtomicBroadcast) StartHealthCheck() {
	const PING_TIMEOUT = 5
	const REQUEST_TIMEOUT = 2

	for {
		time.Sleep(time.Second * time.Duration(PING_TIMEOUT))
		zNode, _ := ab.ZTree.GetLocalMetadata()
		currentPort := zNode.NodePort

		var healthCheck = data.HealthCheck{
			Message:    "ping!",
			PortNumber: currentPort,
		}
		jsonData, _ := json.Marshal(healthCheck)

		servers := strings.Split(zNode.Servers, ",")
		for _, otherPort := range servers {
			if currentPort == otherPort {
				continue
			}
			client := &http.Client{}
			url := fmt.Sprintf(ab.BaseURL + ":" + otherPort + "/")

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Println(err)
				continue
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("X-Sender-Port", zNode.NodePort)

			color.Green("Ping %s", otherPort)

			ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT*time.Second)
			defer cancel()

			req = req.WithContext(ctx)

			// CONNECTION
			resp, err := client.Do(req)
			if resp == nil && err != nil {
				if ctxErr := ctx.Err(); ctxErr == context.DeadlineExceeded {
					color.Red("Timeout Occurred!")
				}
			}
			if err != nil || resp == nil {
				color.Red("Error sending ping to %s", otherPort)

				errorData := data.HealthCheckError{
					Error:     err,
					ErrorPort: otherPort,
					IsWakeup:  false,
				}
				ab.ErrorLeaderChan <- errorData
				continue
			}

			// REPLY
			defer resp.Body.Close()
			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				continue
			}
			var responseObject data.HealthCheck
			json.Unmarshal(resBody, &responseObject)
			if err != nil {
				log.Println(err)
				continue
			}

			// Port number
			color.Green("%s Pong", responseObject.PortNumber)
		}
	}
}

// forwardRequestToLeader for Follower to forward Write Request to Leader
func (ab *AtomicBroadcast) forwardRequestToLeader(r *http.Request) (*http.Response, error) {
	zNode, _ := ab.ZTree.GetLocalMetadata()
	req, _ := http.NewRequest(r.Method, ab.BaseURL+":"+zNode.Leader+r.URL.Path, r.Body)
	req.Header = r.Header
	client := &http.Client{}
	return client.Do(req)
}

// startProposal for Leader to start a 2PC Active Messaging
func (ab *AtomicBroadcast) startProposal(data data.Data) {
	ab.SetProposalState(PROPOSED)

	jsonData, _ := json.Marshal(data)
	zNode, _ := ab.ZTree.GetLocalMetadata()
	portsSlice := strings.Split(zNode.Servers, ",")

	// send Request async
	var wg sync.WaitGroup
	for _, port := range portsSlice {
		if port == zNode.NodePort {
			continue
		}

		wg.Add(1)
		go func(port string) {
			defer wg.Done()

			color.HiBlue("Leader %s proposing to Follower %s", zNode.NodePort, port)
			url := ab.BaseURL + ":" + port + "/proposeWrite"
			_, err := ab.sendRequest(url, "POST", jsonData)
			if err != nil {
				color.Red("Error proposing to follower:", port, "Error:", err)
			}
		}(port)
	}
	wg.Wait()

	// Wait for ACK before committing
	for ab.ProposalState() != ACKNOWLEDGED {
		time.Sleep(time.Second)
	}

	color.HiBlue("Leader %s committing", zNode.NodePort)
	url := ab.BaseURL + ":" + zNode.NodePort + "/writeMetadata"
	_, err := ab.sendRequest(url, "POST", jsonData)
	if err != nil {
		color.Red("Error committing write metadata:", err)
	}
	ab.SetProposalState(COMMITTED)
}

// syncMetadata for new leader to sync its transaction log on joining or restart
func (ab *AtomicBroadcast) syncMetadata() {
	ab.SetSyncState(PREPARED)

	zNode, _ := ab.ZTree.GetLocalMetadata()
	portsSlice := strings.Split(zNode.Servers, ",")

	var metadata ztree.Metadata
	jsonData, _ := json.Marshal(metadata)

	// send Request async
	var wg sync.WaitGroup
	for _, port := range portsSlice {
		if port == zNode.NodePort {
			continue
		}

		wg.Add(1)
		go func(port string) {
			defer wg.Done()

			color.Yellow("%s send syncRequest to %s", zNode.NodePort, port)
			url := ab.BaseURL + ":" + port + "/syncRequest"
			_, err := ab.sendRequest(url, "POST", jsonData)
			if err != nil {
				color.Red("Error syncRequest to %s:", port, "Error:", err)
			}
		}(port)
	}
	wg.Wait()

	// Wait for ACK before request Metadata
	for ab.SyncState() != ACKED {
		time.Sleep(time.Second)
	}

	highestZNodeId, _ := ab.ZTree.GetHighestZNodeId()
	metadata = ztree.Metadata{
		NodeId: highestZNodeId,
	}
	jsonData, _ = json.Marshal(metadata)

	for _, port := range portsSlice {
		if port == zNode.NodePort {
			continue
		}

		color.Yellow("%s requestMetadata from %s", zNode.NodePort, port)
		url := ab.BaseURL + ":" + port + "/requestMetadata"
		_, err := ab.sendRequest(url, "POST", jsonData)
		if err != nil {
			color.Red("Error requestMetadata to:", port, "Error:", err)
		}
	}

	color.Yellow("%s finished syncing", zNode.NodePort)
	ab.SetSyncState(SYNCED)
}

// WakeupLeaderElection for new ZooWeeper server to declare itself when joining or restart
func (ab *AtomicBroadcast) WakeupLeaderElection(port int) {
	for {
		select {
		case <-time.After(time.Second * 2):
			// On wake-up, start leader-election
			data := data.HealthCheckError{
				Error:     nil,
				ErrorPort: fmt.Sprintf("%d", port),
				IsWakeup:  true,
			}
			ab.ErrorLeaderChan <- data
			return
		}
	}
}

// ListenForLeaderElection when there is data from ErrorLeaderChan due to TimeOut
func (ab *AtomicBroadcast) ListenForLeaderElection(port int, leader int) {
	for {
		select {
		case errorData := <-ab.ErrorLeaderChan:
			errorPortNumber, _ := strconv.Atoi(errorData.ErrorPort)
			if errorPortNumber == leader || errorData.IsWakeup {
				if errorData.IsWakeup {
					color.Cyan("%d joining, starting election", port)
				} else if errorPortNumber == leader {
					color.Cyan("Healthcheck timeout for %d, starting election", errorPortNumber)
				}
				ab.startLeaderElection(port)
			}
		}
	}
}

// startLeaderElection starts Bully by calling its own server handler with information of the port information.
func (ab *AtomicBroadcast) startLeaderElection(currentPort int) {
	client := &http.Client{}
	portURL := fmt.Sprintf("%d", currentPort)

	url := fmt.Sprintf(ab.BaseURL + ":" + portURL + "/electLeader")
	var electMessage = data.ElectLeaderRequest{
		IncomingPort: fmt.Sprintf("%d", currentPort),
	}
	jsonData, _ := json.Marshal(electMessage)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

// declareLeaderRequest sends <declare-leader> message to all Followers
func (ab *AtomicBroadcast) declareLeaderRequest(portStr string, allServers []string) {
	for _, outgoingPort := range allServers {
		client := &http.Client{}
		portURL := fmt.Sprintf("%s", outgoingPort)

		url := fmt.Sprintf(ab.BaseURL + ":" + portURL + "/declareLeaderReceive")
		var electMessage = data.DeclareLeaderRequest{
			IncomingPort: portStr,
		}
		jsonData, _ := json.Marshal(electMessage)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			continue
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		color.Cyan("%s declare Leader to %s", portStr, outgoingPort)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
	}
}
