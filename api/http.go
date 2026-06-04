package api

import (
	"encoding/json"
	"net/http"

	"github.com/PratikkJadhav/MiniObs/query"
	"github.com/PratikkJadhav/MiniObs/storage"
	"github.com/go-chi/chi/v5"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(store *storage.Store) http.Handler {
	r := chi.NewRouter()

	r.Use(withCORS)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "ui/index.html")
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	r.Get("/api/services", func(w http.ResponseWriter, r *http.Request) {
		services := store.GetServices()
		jsonData, _ := json.Marshal(services)

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})
	r.Get("/api/traces", func(w http.ResponseWriter, r *http.Request) {
		service := r.URL.Query().Get("service")
		var summaries []storage.TraceSummary
		if service == "" {
			for _, svc := range store.GetServices() {
				s, err := store.GetTraceSummaries(svc)
				if err != nil {
					http.Error(w, "failed to read traces", http.StatusInternalServerError)
					return
				}
				summaries = append(summaries, s...)
			}
		} else {
			var err error
			summaries, err = store.GetTraceSummaries(service)
			if err != nil {
				http.Error(w, "failed to read traces", http.StatusInternalServerError)
				return
			}
		}
		result := map[string]interface{}{"traces": summaries}
		json.NewEncoder(w).Encode(result)
	})
	r.Get("/api/traces/{traceID}", func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceID")
		spans, err := store.GetTraceByID(traceID)
		if err != nil {
			http.Error(w, "failed to read trace", http.StatusInternalServerError)
			return
		}
		jsonData, _ := json.Marshal(spans)

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	r.Get("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		service := r.URL.Query().Get("service")
		if service == "" {
			http.Error(w, "service param required", http.StatusBadRequest)
			return
		}
		p50, p95, p99, errorRate, err := query.ComputeMetrics(store, service)
		if err != nil {
			http.Error(w, "failed to read trace", http.StatusInternalServerError)
			return
		}
		data := map[string]float64{
			"p50":        p50,
			"p95":        p95,
			"p99":        p99,
			"error_rate": errorRate,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	return r
}
