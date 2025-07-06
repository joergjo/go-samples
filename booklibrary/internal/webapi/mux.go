package webapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joergjo/go-samples/booklibrary/internal/log"
	"github.com/joergjo/go-samples/booklibrary/internal/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMux creates a new route multiplexer for all endpoints offered the BookLibrary API and all required middleware enabled.
func NewMux(crud model.CrudService) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Heartbeat("/healthz/live"))
	r.Get("/healthz/ready", readyHandler(crud))
	r.Mount("/api/books", NewResource(crud))
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())
	return r
}

func readyHandler(crud model.CrudService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := crud.Ping(ctx); err != nil {
			slog.Error("ping database", log.ErrorKey, err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	}
}
