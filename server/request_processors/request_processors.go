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

func (rp *RequestProcessor) Routes() http.Handler {
	mux := chi.NewRouter()

	// middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(rp.Zab.EnableCORS)

	// routes
	mux.Get("/", rp.Zab.Ping)
	mux.Post("/score", rp.Zab.AddScore)
	mux.Post("/metadata", rp.Zab.UpdateMetaData)

	mux.Get("/metadata", rp.Zab.GetAllMetadata)
	mux.Post("/scoreExists/{leaderServer}", rp.Zab.DoesScoreExist)
	mux.Delete("/score/{leaderServer}", rp.Zab.DeleteScore)

	return mux
}
