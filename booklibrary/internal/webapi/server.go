package webapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/joergjo/go-samples/booklibrary/internal/log"
	"github.com/joergjo/go-samples/booklibrary/internal/model"
)

// NewServer creates a new HTTP server with the given handler and port.
func NewServer(crud model.CrudService, port int) *http.Server {
	mux := NewMux(crud)
	s := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	return &s
}

// Shutdown gracefully shuts down the server with a timeout of 30 seconds.
func Shutdown(ctx context.Context, s *http.Server, done chan struct{}) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigch
	slog.Warn(fmt.Sprintf("got signal %v", sig))

	childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.Shutdown(childCtx); err != nil {
		slog.Error("shutdown", log.ErrorKey, err)
	}
	close(done)
}
