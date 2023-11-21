// ReadOps and WriteOps have Self-reference to parent - ref: https://stackoverflow.com/questions/27918208/go-get-parent-struct

package zooweeper

import (
	"container/heap"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	ztree "github.com/tnbl265/zooweeper/database"
	"github.com/tnbl265/zooweeper/database/handlers"
	"github.com/tnbl265/zooweeper/database/models"
)

type ProposalState string

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
	ZTree    ztree.ZooWeeperDatabaseRepo

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

	ab.proposalState = COMMITTED

	ab.ErrorLeaderChan = make(chan models.HealthCheckError)

	ab.pq = make(PriorityQueue, 0)
	heap.Init(&ab.pq)

	ab.ZTree.InitializeDB()
	return ab
}

func (ab *AtomicBroadcast) Ping(portStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload models.HealthCheck
		err := ab.readJSON(w, r, &requestPayload)
		if err != nil {
			ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		color.Magenta("Received Healthcheck from Port:%s , Message:%s \n", requestPayload.PortNumber, requestPayload.Message)

		payload := models.HealthCheck{
			Message:    "pong",
			PortNumber: portStr,
		}

		_ = ab.writeJSON(w, http.StatusOK, payload)
	}
}
func (ab *AtomicBroadcast) ElectLeader(portStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload models.ElectLeaderRequest
		err := ab.readJSON(w, r, &requestPayload)
		if err != nil {
			ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		color.Magenta("Received election message from Port:%s \n", requestPayload.IncomingPort)

		incomingPortNumber, _ := strconv.Atoi(requestPayload.IncomingPort)
		currentPortNumber, _ := strconv.Atoi(portStr)

		payload := models.ElectLeaderResponse{
			IsSuccess: strconv.FormatBool(incomingPortNumber > currentPortNumber),
		}

		_ = ab.writeJSON(w, http.StatusOK, payload)
	}
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

func (ab *AtomicBroadcast) readJSON2(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1024 * 1024 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
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
			_ = ab.makeExternalRequest(nil, url, "POST", jsonData)
		}
	}

	// Wait for ACK before committing
	for ab.ProposalState() != ACKNOWLEDGED {
		time.Sleep(time.Second)
	}
	log.Println("Leader committing")
	url := ab.BaseURL + ":" + zNode.NodeIp + "/writeMetadata"
	_ = ab.makeExternalRequest(nil, url, "POST", jsonData)
	ab.SetProposalState(COMMITTED)
}
