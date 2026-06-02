package api

import (
	"net/http"

	"github.com/PratikkJadhav/MiniObs/storage"
	"github.com/go-chi/chi/v5"
)

func NewRouter(store *storage.Store) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/services", func(w http.ResponseWriter, r *http.Request) {
		services := store.GetServices()

	})
	r.Get("/api/traces", store.GetTraceIDs)
	r.Get("/api/traces/{traceID}", store.GetTraceByID)

	return r
}
