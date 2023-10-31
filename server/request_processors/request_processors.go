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

func (rp *RequestProcessor) Routes() http.Handler {
	mux := chi.NewRouter()

	// middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(rp.Zab.EnableCORS)

	// routes
	mux.Get("/", rp.Zab.Ping)
	mux.Get("/metadata", rp.Zab.Read.GetAllMetadata)
	mux.Post("/scoreExists/{leaderServer}", rp.Zab.Read.DoesScoreExist)

	mux.Post("/score", rp.Zab.Write.AddScore)
	mux.Post("/metadata", rp.Zab.Write.UpdateMetaData)
	mux.Delete("/score/{leaderServer}", rp.Zab.Write.DeleteScore)

	return mux
}
