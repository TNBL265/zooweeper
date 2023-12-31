// Package request_processors implements the Request Processor component for our ZooWeeper. We make use of HTTP Protocol
//
// 1. All Write Request are forwarded to the Leader while Read Request are done locally in each server
// 2. Write Request are captured as Transaction with Timestamp field (generated by Kafka broker) as id
// 3. QueueMiddleware will order Transaction by Timestamp using a Priority Queue helps guarantee
//   - FIFO Client order: all Transaction is ordered by Timestamp generated by client (Kafka broker)
//   - Linearization Write: Transaction is only processed by the Leader using a classic 2PC Active Messaging
//
// 4. WriteOpsMiddleware:
//   - Follower: forward request to Leader
//   - Leader: start write proposal for as a classic two-phase commit
//
// 5. We also define other internal requests for some Distributed System features:
// - Proposal Request for Data Synchronization when all ZooWeeper servers are healthy
// - Leader Election Request: Distributed Coordination
// - Data Sync Request for Data Synchronization when a ZooWeeper server joined or restarted, ensuring Fault Tolerance
//
// Reference: Active Messaging in https://zookeeper.apache.org/doc/current/zookeeperInternals.html#sc_activeMessaging

package request_processors

import (
	"container/heap"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tnbl265/zooweeper/zab"
)

type RequestProcessor struct {
	Zab *zab.AtomicBroadcast
	pq  PriorityQueue
}

func NewRequestProcessor(dbPath string) *RequestProcessor {
	rp := &RequestProcessor{}
	rp.Zab = zab.NewAtomicBroadcast(dbPath)
	rp.pq = make(PriorityQueue, 0)
	heap.Init(&rp.pq)

	return rp
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
		r.Use(rp.QueueMiddleware)
		r.Use(rp.WriteOpsMiddleware)

		r.Post("/metadata", rp.Zab.Write.UpdateMetadata)
	})

	// Proposal Request
	mux.Group(func(r chi.Router) {
		r.Post("/proposeWrite", rp.Zab.Proposal.ProposeWrite)
		r.Post("/acknowledgeProposal", rp.Zab.Proposal.AcknowledgeProposal)
		r.Post("/commitWrite", rp.Zab.Proposal.CommitWrite)
		r.Post("/writeMetadata", rp.Zab.Write.WriteMetadata)
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
