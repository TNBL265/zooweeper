// ReadOps and WriteOps have Self-reference to parent - ref: https://stackoverflow.com/questions/27918208/go-get-parent-struct

package zooweeper

import (
	"bytes"
	"container/heap"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	ztree "github.com/tnbl265/zooweeper/database"
	"github.com/tnbl265/zooweeper/database/handlers"
	"github.com/tnbl265/zooweeper/database/models"
)

type ProposalState string

var err error

const (
	COMMITTED    ProposalState = "COMMITTED"
	PROPOSED     ProposalState = "PROPOSED"
	ACKNOWLEDGED ProposalState = "ACKNOWLEDGED"
)

type AtomicBroadcast struct {
	BaseURL string

	Read     ReadOps
	Write    WriteOps
	Proposal ProposalOps
	Election ElectionOps

	ZTree ztree.ZooWeeperDatabaseRepo

	// Proposal
	ackCounter    int
	proposalState ProposalState
	proposalMu    sync.Mutex

	pq PriorityQueue

	ErrorLeaderChan chan models.HealthCheckError
}

func (ab *AtomicBroadcast) AckCounter() int {
	ab.proposalMu.Lock()
	defer ab.proposalMu.Unlock()
	return ab.ackCounter
}

func (ab *AtomicBroadcast) SetAckCounter(ackCounter int) {
	ab.proposalMu.Lock()
	defer ab.proposalMu.Unlock()
	ab.ackCounter = ackCounter
}

func (ab *AtomicBroadcast) ProposalState() ProposalState {
	ab.proposalMu.Lock()
	defer ab.proposalMu.Unlock()
	return ab.proposalState
}

func (ab *AtomicBroadcast) SetProposalState(proposalState ProposalState) {
	ab.proposalMu.Lock()
	defer ab.proposalMu.Unlock()
	ab.proposalState = proposalState
	log.Printf("Set ProposalState to %s\n", proposalState)
}

func NewAtomicBroadcast(dbPath string) *AtomicBroadcast {
	ab := &AtomicBroadcast{}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	ab.BaseURL = baseURL

	// Connect to the Database
	log.Println("Connecting to", dbPath)
	db, err := ab.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	ab.ZTree = &handlers.ZTree{DB: db}
	ab.Read.ab = ab
	ab.Write.ab = ab
	ab.Proposal.ab = ab
	ab.Election.ab = ab

	ab.proposalState = COMMITTED

	ab.ErrorLeaderChan = make(chan models.HealthCheckError)

	ab.pq = make(PriorityQueue, 0)
	heap.Init(&ab.pq)

	ab.ZTree.InitializeDB()
	return ab
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

func (ab *AtomicBroadcast) CreateMetadata(w http.ResponseWriter, r *http.Request) models.Data {
	var requestPayload models.Data

	err := ab.readJSON(w, r, &requestPayload)
	if err != nil {
		ab.errorJSON(w, err, http.StatusBadRequest)
		return models.Data{}
	}
	data := models.Data{
		Timestamp:   requestPayload.Timestamp,
		Metadata:    requestPayload.Metadata,
		GameResults: requestPayload.GameResults,
	}
	return data
}

func (ab *AtomicBroadcast) forwardRequestToLeader(r *http.Request) (*http.Response, error) {
	zNode, _ := ab.ZTree.GetLocalMetadata()
	req, _ := http.NewRequest(r.Method, ab.BaseURL+":"+zNode.Leader+r.URL.Path, r.Body)
	req.Header = r.Header
	client := &http.Client{}
	return client.Do(req)
}

func (ab *AtomicBroadcast) startProposal(data models.Data) {
	ab.SetProposalState(PROPOSED)

	jsonData, _ := json.Marshal(data)
	zNode, _ := ab.ZTree.GetLocalMetadata()
	portsSlice := strings.Split(zNode.Servers, ",")
	for _, port := range portsSlice {
		if port != zNode.NodeIp {
			log.Println("Proposing to Follower:", port)
			url := ab.BaseURL + ":" + port + "/proposeWrite"
			_, err := ab.makeExternalRequest(nil, url, "POST", jsonData)
			if err != nil {
				log.Println("Error proposing to follower:", port, "Error:", err)
				continue
			}
		}
	}

	// Wait for ACK before committing
	for ab.ProposalState() != ACKNOWLEDGED {
		time.Sleep(time.Second)
	}

	log.Println("Leader committing")
	url := ab.BaseURL + ":" + zNode.NodeIp + "/writeMetadata"
	_, err := ab.makeExternalRequest(nil, url, "POST", jsonData)
	if err != nil {
		log.Println("Error committing write metadata:", err)
	}
	ab.SetProposalState(COMMITTED)
}

func (ab *AtomicBroadcast) WakeupLeaderElection(port int) {
	for {
		select {
		case <-time.After(time.Second * 2):
			// On wake-up, start leader-election
			data := models.HealthCheckError{
				Error:     nil,
				ErrorPort: fmt.Sprintf("%d", port),
				IsWakeup:  true,
			}
			ab.ErrorLeaderChan <- data
			return
		}
	}
}
func (ab *AtomicBroadcast) ListenForLeaderElection(port int, leader int) {
	for {
		select {
		case errorData := <-ab.ErrorLeaderChan:
			errorPortNumber, _ := strconv.Atoi(errorData.ErrorPort)
			if errorPortNumber == leader || errorData.IsWakeup {
				if errorPortNumber == leader {
					color.Magenta("Error from port %d healthcheck! Starting leader election here...", errorPortNumber)
				} else if errorData.IsWakeup {
					color.Magenta("Port %d just woke up! Starting leader election here...", port)
				}

				ab.startLeaderElection(port)
			}
		}
	}
}

// Starts the leader election by calling its own server handler with information of the port information.
func (ab *AtomicBroadcast) startLeaderElection(currentPort int) {

	client := &http.Client{}
	portURL := fmt.Sprintf("%d", currentPort)

	url := fmt.Sprintf(ab.BaseURL + ":" + portURL + "/electLeader")
	var electMessage = models.ElectLeaderRequest{
		IncomingPort: fmt.Sprintf("%d", currentPort),
	}
	jsonData, _ := json.Marshal(electMessage)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
}
