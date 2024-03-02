package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"log/slog"

	"github.com/joergjo/go-samples/booklibrary/internal/config"
	"github.com/joergjo/go-samples/booklibrary/internal/http/router"
	"github.com/joergjo/go-samples/booklibrary/internal/http/server"
	"github.com/joergjo/go-samples/booklibrary/internal/log"
	"github.com/joergjo/go-samples/booklibrary/internal/mongo"
)

func main() {
	s := configure()
	slog.SetDefault(log.New(os.Stdout, s.Debug))

	crud, err := newCrudService(s.MongoURI, s.Db, s.Collection)
	if err != nil {
		slog.Error("creating book service", log.ErrorKey, err)
		os.Exit(1)
	}
	defer func() {
		if err := crud.Close(context.Background()); err != nil {
			slog.Error("closing database connection", log.ErrorKey, err)
		} else {
			slog.Info("closed database connection")
		}
	}()

	mux := router.NewMux(crud)
	srv := server.New(mux, s.Port)
	done := make(chan struct{})
	go server.Shutdown(context.Background(), srv, done)

	slog.Info(fmt.Sprintf("server starting, listening on 0.0.0.0:%d", s.Port))
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", log.ErrorKey, err)
		os.Exit(1)
	}
	slog.Info("waiting for shut down to complete")
	<-done
	slog.Info("server has shut down")
}

func configure() config.Settings {
	var s config.Settings
	mongoURI := config.GetEnvString("BOOKLIBRARY_MONGOURI", "mongodb://localhost")
	port := config.GetEnvInt("BOOKLIBRARY_PORT", 8000)
	db := config.GetEnvString("BOOKLIBRARY_DB", "library_database")
	coll := config.GetEnvString("BOOKLIBRARY_COLLECTION", "books")
	debug := config.GetEnvBool("BOOKLIBRARY_DEBUG", false)

	flag.IntVar(&(s.Port), "port", port, "HTTP port to listen on")
	flag.StringVar(&(s.MongoURI), "mongoURI", mongoURI, "MongoDB URI to connect to")
	flag.StringVar(&(s.Db), "db", db, "MongoDB database")
	flag.StringVar(&(s.Collection), "collection", coll, "MongoDB collection")
	flag.BoolVar(&(s.Debug), "debug", debug, "Enable debug logging")
	flag.Parse()
	return s
}

func newCrudService(uri, db, coll string) (*mongo.CrudService, error) {
	slog.Debug(fmt.Sprintf("connecting to MongoDB at %q", uri))
	crud, err := mongo.NewCrudService(uri, db, coll)
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("connected to MongoDB at %q", uri))
	return crud, nil
}
