package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"log/slog"

	"github.com/joergjo/go-samples/booklibrary/internal/config"
	"github.com/joergjo/go-samples/booklibrary/internal/log"
	"github.com/joergjo/go-samples/booklibrary/internal/mongo"
	"github.com/joergjo/go-samples/booklibrary/internal/webapi"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	s := configure()
	slog.SetDefault(log.New(os.Stdout, s.Debug))

	slog.Info("booklibrary-api", "version", version, "commit", commit, "date", date, "builtBy", builtBy, "goVersion", runtime.Version())
	if s.Debug {
		slog.Warn("debug logging enabled")
	}

	os.Exit(run(s))
}

func run(s config.Settings) int {
	crud, err := newCrudService(s.MongoURI, s.Db, s.Collection)
	if err != nil {
		slog.Error("creating book service", log.ErrorKey, err)
		return 1
	}
	defer func() {
		slog.Info("closing database connection")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := crud.Close(ctx); err != nil {
			slog.Error("closing database connection", log.ErrorKey, err)
		} else {
			slog.Info("closed database connection")
		}
	}()

	srv := webapi.NewServer(crud, s.Port)

	errC := make(chan error, 1)
	go func() {
		slog.Info("starting server", log.AddrKey, srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errC <- err
		}
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	var exit int
	select {
	case err := <-errC:
		slog.Error("server error", log.ErrorKey, err)
		exit = 1
	case sig := <-sigC:
		slog.Warn("received signal, shutting down", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		slog.Info("waiting for server to shut down")
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutting down server", "error", err)
			if err := srv.Close(); err != nil {
				slog.Error("forcefully closing server", "error", err)
			}
		}
		slog.Info("server has shut down")
	}

	return exit
}

func configure() config.Settings {
	var s config.Settings
	mongoURI := config.GetEnvString("BOOKLIBRARY_MONGOURI", "mongodb://localhost/?timeoutMS=0")
	port := config.GetEnvInt("BOOKLIBRARY_PORT", 8000)
	db := config.GetEnvString("BOOKLIBRARY_DB", "library_database")
	coll := config.GetEnvString("BOOKLIBRARY_COLLECTION", "books")
	debug := config.GetEnvBool("BOOKLIBRARY_DEBUG", false)

	flag.IntVar(&s.Port, "port", port, "HTTP port to listen on")
	flag.StringVar(&s.MongoURI, "mongoURI", mongoURI, "MongoDB URI to connect to")
	flag.StringVar(&s.Db, "db", db, "MongoDB database")
	flag.StringVar(&s.Collection, "collection", coll, "MongoDB collection")
	flag.BoolVar(&s.Debug, "debug", debug, "Enable debug logging")
	flag.Parse()
	return s
}

func newCrudService(uri, db, coll string) (*mongo.CrudService, error) {
	slog.Debug("connecting to MongoDB", log.MongoURIKey, uri)
	crud, err := mongo.NewCrudService(uri, db, coll)
	if err != nil {
		return nil, err
	}
	slog.Debug("connected to MongoDB", log.MongoURIKey, uri)
	return crud, nil
}
