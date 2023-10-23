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
	mux.Get("/gameResults", app.GetGameResults)

	return mux
}
