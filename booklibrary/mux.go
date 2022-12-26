package booklibrary

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMux(crud CrudService) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Heartbeat("/healthz/live"))
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())
	r.Mount("/api/books", Routes(crud))
	r.Get("/healthz/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := crud.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})
	return r
}
