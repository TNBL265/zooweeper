package zooweeper

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
		r.Post("/", rp.Zab.Ping(portStr))
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
		r.Post("/electLeader", rp.Zab.SelfElectLeaderRequest(portStr))
		r.Post("/declareLeaderReceive", rp.Zab.DeclareLeaderReceive(portStr))
	})

	return mux
}
