// Package request_processors implements the Request Processor for our ZooWeeper components. We make use of HTTP Protocol
//
// 1. Read Request will be handled locally (no middleware)
// 2. Write Request will be forwarded to Leader node
// - QueueMiddleware helps to establish a form of Total Order to guarantee FIFO client order and Linearization writes
// - WriteMiddleware:
//   - Follower: forward request to Leader
//   - Leader: start write proposal for as a classic two-phase commit
//
// 3. We also define other internal requests for some Distributed System features:
// - Proposal Request for Data Synchronization when all ZooWeeper servers are healthy
// - Leader Election Request: Distributed Coordination
// - Data Sync Request for Data Synchronization when a ZooWeeper server joined or restarted
//
// Reference: Active Messaging in https://zookeeper.apache.org/doc/current/zookeeperInternals.html
package request_processors

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	zab "github.com/tnbl265/zooweeper/zab"
)

type RequestProcessor struct {
	Zab zab.AtomicBroadcast
}

func NewRequestProcessor(dbPath string) *RequestProcessor {
	return &RequestProcessor{
		Zab: *zab.NewAtomicBroadcast(dbPath),
	}
}

func (rp *RequestProcessor) Routes(portStr string) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(rp.Zab.EnableCORS)
	mux.Use(middleware.WithValue("portStr", portStr))

	// Read Request
	mux.Group(func(r chi.Router) {
		r.Get("/metadata", rp.Zab.Read.GetAllMetadata)
	})

	// Write Request
	mux.Group(func(r chi.Router) {
		r.Use(rp.Zab.QueueMiddleware)
		r.Use(rp.Zab.Write.WriteOpsMiddleware)

		r.Post("/metadata", rp.Zab.Write.UpdateMetaData)
	})

	// Proposal Request
	mux.Group(func(r chi.Router) {
		r.Post("/proposeWrite", rp.Zab.Proposal.ProposeWrite)
		r.Post("/acknowledgeProposal", rp.Zab.Proposal.AcknowledgeProposal)
		r.Post("/commitWrite", rp.Zab.Proposal.CommitWrite)
		r.Post("/writeMetadata", rp.Zab.Write.WriteMetaData)
	})

	// Leader Election Request
	mux.Group(func(r chi.Router) {
		r.Post("/", rp.Zab.Election.Ping(portStr))
		r.Post("/electLeader", rp.Zab.Election.SelfElectLeaderRequest(portStr))
		r.Post("/declareLeaderReceive", rp.Zab.Election.DeclareLeaderReceive())
	})

	// Data Sync Request
	mux.Group(func(r chi.Router) {
		r.Post("/syncRequest", rp.Zab.Sync.SyncRequestHandler)
		r.Post("/syncResponse", rp.Zab.Sync.SyncResponseHandler)
		r.Post("/requestMetadata", rp.Zab.Sync.RequestMetadataHandler)
		r.Post("/updateMetadata", rp.Zab.Sync.UpdateMetadataHandler)
	})

	return mux
}
