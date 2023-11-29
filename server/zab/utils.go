package zab

import (
	"bytes"
	"container/heap"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/tnbl265/zooweeper/request_processors/data"
	"github.com/tnbl265/zooweeper/ztree"
	"io"
	"log"
	"net/http"
	"os"
)

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
	color.HiRed("Set ProposalState to %s\n", proposalState)
}

func (ab *AtomicBroadcast) SyncCounter() int {
	ab.syncMu.Lock()
	defer ab.syncMu.Unlock()
	return ab.syncCounter
}

func (ab *AtomicBroadcast) SetSyncCounter(syncCounter int) {
	ab.syncMu.Lock()
	defer ab.syncMu.Unlock()
	ab.syncCounter = syncCounter
}

func (ab *AtomicBroadcast) SyncState() SyncState {
	ab.syncMu.Lock()
	defer ab.syncMu.Unlock()
	return ab.syncState
}

func (ab *AtomicBroadcast) SetSyncState(syncState SyncState) {
	ab.syncMu.Lock()
	defer ab.syncMu.Unlock()
	ab.syncState = syncState
	color.HiRed("Set SyncState to %s\n", syncState)
}

// NewAtomicBroadcast self-reference to parent - ref: https://stackoverflow.com/questions/27918208/go-get-parent-struct
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
	ab.ZTree = &ztree.ZTree{DB: db}
	ab.Read.ab = ab
	ab.Write.ab = ab
	ab.Proposal.ab = ab
	ab.Election.ab = ab
	ab.Sync.ab = ab

	ab.proposalState = COMMITTED

	ab.ErrorLeaderChan = make(chan data.HealthCheckError)

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

func (ab *AtomicBroadcast) CreateMetadataFromPayload(w http.ResponseWriter, r *http.Request) data.Data {
	var requestPayload data.Data

	ab.readJSON(w, r, &requestPayload)

	data := data.Data{
		Timestamp:   requestPayload.Timestamp,
		Metadata:    requestPayload.Metadata,
		GameResults: requestPayload.GameResults,
	}
	return data
}

func (ab *AtomicBroadcast) EnableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "Options" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (ab *AtomicBroadcast) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(out)

	if err != nil {
		return err
	}

	return nil
}
func (ab *AtomicBroadcast) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1024 * 1024 // 1mb
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

func (ab *AtomicBroadcast) sendRequest(incomingUrl string, method string, jsonData []byte) (*http.Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf(incomingUrl)

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonData))

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	zNode, err := ab.ZTree.GetLocalMetadata()
	req.Header.Add("X-Sender-Port", zNode.NodePort)

	res, err := client.Do(req)
	if err != nil {
		//log.Println("Error sending request:", err)
		return nil, err
	}

	return res, nil
}
