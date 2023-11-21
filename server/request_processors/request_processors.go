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
		r.Post("/electLeader", rp.Zab.ElectLeader(portStr))
		r.Get("/metadata", rp.Zab.Read.GetAllMetadata)
		r.Post("/scoreExists/{leader}", rp.Zab.Read.DoesScoreExist)
	})

	// Write Request
	mux.Group(func(r chi.Router) {
		r.Use(rp.Zab.QueueMiddleware)
		r.Use(rp.Zab.Write.WriteOpsMiddleware)

		//r.Post("/score", rp.Zab.Write.AddScore)
		r.Post("/metadata", rp.Zab.Write.UpdateMetaData)
		r.Delete("/score/{leader}", rp.Zab.Write.DeleteScore)
	})

	// Proposal Request
	mux.Group(func(r chi.Router) {
		mux.Post("/proposeWrite", rp.Zab.Proposal.ProposeWrite)
		mux.Post("/acknowledgeProposal", rp.Zab.Proposal.AcknowledgeProposal)
		mux.Post("/commitWrite", rp.Zab.Proposal.CommitWrite)
		r.Post("/writeMetadata", rp.Zab.Write.WriteMetaData)
	})

	return mux
}
