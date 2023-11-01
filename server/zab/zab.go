// ReadOps and WriteOps have Self-reference to parent - ref: https://stackoverflow.com/questions/27918208/go-get-parent-struct

package zooweeper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	ztree "github.com/tnbl265/zooweeper/database"
	"github.com/tnbl265/zooweeper/database/handlers"
	"github.com/tnbl265/zooweeper/database/models"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ProposalState string

const (
	COMMITTED    ProposalState = "COMMITTED"
	PROPOSED     ProposalState = "PROPOSED"
	ACKNOWLEDGED ProposalState = "ACKNOWLEDGED"
)

type AtomicBroadcast struct {
	Read     ReadOps
	Write    WriteOps
	Proposal ProposalOps
	ZTree    ztree.ZooWeeperDatabaseRepo

	// Proposal
	ackCounter    int
	proposalState ProposalState
	proposalMu    sync.Mutex
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

	ab.ZTree.InitializeDB()
	return ab
}

func (ab *AtomicBroadcast) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
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

func (ab *AtomicBroadcast) CreateMetadata(w http.ResponseWriter, r *http.Request) models.Metadata {
	var requestPayload models.Metadata

	err := ab.readJSON(w, r, &requestPayload)
	if err != nil {
		ab.errorJSON(w, err, http.StatusBadRequest)
		return models.Metadata{}
	}
	metadata := models.Metadata{
		SenderIp:   requestPayload.SenderIp,
		ReceiverIp: requestPayload.ReceiverIp,
		Attempts:   requestPayload.Attempts,
		Timestamp:  requestPayload.Timestamp,
	}
	return metadata
}

func (ab *AtomicBroadcast) startProposal(metadata models.Metadata) {
	ab.SetProposalState(PROPOSED)

	jsonData, _ := json.Marshal(metadata)

	zNode, _ := ab.ZTree.GetLocalMetadata()
	portsSlice := strings.Split(zNode.Servers, ",")
	for _, port := range portsSlice {
		if port != zNode.NodeIp {
			log.Println("Proposing to Follower:", port)
			url := "http://localhost:" + port + "/proposeWrite"
			_ = ab.makeExternalRequest(nil, url, "POST", jsonData)
		}
	}

	// Wait for ACK before committing
	for ab.ProposalState() != ACKNOWLEDGED {
		time.Sleep(time.Second)
	}
	log.Println("Leader committing")
	url := "http://localhost:" + zNode.NodeIp + "/writeMetadata"
	_ = ab.makeExternalRequest(nil, url, "POST", jsonData)
	ab.SetProposalState(COMMITTED)
}
