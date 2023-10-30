package zooweeper

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Application) Routes() http.Handler {
	mux := chi.NewRouter()

	// middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	// routes
	mux.Get("/", app.Ping)
	mux.Post("/score", app.AddScore)
	mux.Post("/metadata", app.UpdateMetaData)

	mux.Get("/metadata", app.GetAllMetadata)
	mux.Post("/scoreExists/{leaderServer}", app.doesScoreExist)
	mux.Delete("/score/{leaderServer}", app.DeleteScore)

	return mux
}
