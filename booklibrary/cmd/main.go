package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"github.com/joergjo/go-samples/booklibrary/mongo"
	"golang.org/x/exp/slog"
)

type appConfig struct {
	port       int
	mongoURI   string
	db         string
	collection string
	debug      bool
}

func main() {
	conf := config()
	slog.SetDefault(newLogger(os.Stdout, conf.debug))

	crud, err := newCrudService(conf.mongoURI, conf.db, conf.collection)
	if err != nil {
		slog.Error("Fatal error creating book service", err)
		os.Exit(1)
	}
	defer func() {
		if err := crud.Close(context.Background()); err != nil {
			slog.Error("Error closing database connection", err)
		} else {
			slog.Info("Closed database connection")
		}
	}()

	r := booklibrary.NewMux(crud)
	srv := newServer(r, conf.port)
	done := make(chan struct{})
	go shutdown(context.Background(), srv, done)

	slog.Info(fmt.Sprintf("Server starting, listening on 0.0.0.0:%d", conf.port))
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("Server error", err)
	}
	slog.Info("Waiting for shut down to complete")
	<-done
	slog.Info("Server has shut down")
}

func shutdown(ctx context.Context, s *http.Server, done chan struct{}) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigch
	slog.Warn(fmt.Sprintf("Got signal %v", sig))

	childCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.Shutdown(childCtx); err != nil {
		slog.Error("Error during shutdown", err)
	}
	close(done)
}

func config() appConfig {
	var a appConfig
	mongoURI := getEnvString("BOOKLIBRARY_MONGOURI", "mongodb://localhost")
	port := getEnvInt("BOOKLIBRARY_PORT", 8000)
	db := getEnvString("BOOKLIBRARY_DB", "library_database")
	coll := getEnvString("BOOKLIBRARY_COLLECTION", "books")
	debug := getEnvBool("BOOKLIBRARY_DEBUG", false)

	flag.IntVar(&(a.port), "port", port, "HTTP port to listen on")
	flag.StringVar(&(a.mongoURI), "mongoURI", mongoURI, "MongoDB URI to connect to")
	flag.StringVar(&(a.db), "db", db, "MongoDB database")
	flag.StringVar(&(a.collection), "collection", coll, "MongoDB collection")
	flag.BoolVar(&(a.debug), "debug", debug, "Enable debug logging")
	flag.Parse()
	return a
}

func newLogger(w io.Writer, debug bool) *slog.Logger {
	opts := slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(a.Key, a.Value.Time().UTC())
			}
			return a
		},
	}
	if debug {
		opts.Level = slog.LevelDebug
	}

	return slog.New(opts.NewTextHandler(w))
}

func newCrudService(uri, db, coll string) (*mongo.CrudService, error) {
	slog.Debug(fmt.Sprintf("Connecting to MongoDB at %q", uri))
	crud, err := mongo.NewCrudService(uri, db, coll)
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("Connected to MongoDB at %q", uri))
	return crud, nil
}

func newServer(h http.Handler, port int) *http.Server {
	s := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      h,
	}
	return &s
}

func getEnvString(name, value string) string {
	env, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	return env
}

func getEnvInt(name string, value int) int {
	envStr, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	env, err := strconv.Atoi(envStr)
	if err != nil {
		return value
	}
	return env
}

func getEnvBool(name string, value bool) bool {
	envStr, ok := os.LookupEnv(name)
	if !ok {
		return value
	}
	env, err := strconv.ParseBool(envStr)
	if err != nil {
		return value
	}
	return env
}
