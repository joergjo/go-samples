package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	api "github.com/joergjo/go-samples/booklibrary/http"
	"github.com/joergjo/go-samples/booklibrary/mongo"
)

var appConfig = struct {
	port       int
	mongoURI   string
	db         string
	collection string
}{}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[booklibrary-api] ")

	config()
	store, err := newStorage()
	if err != nil {
		log.Fatalf("Fatal error creating storage implementation: %v\n", err)
	}
	srv := newServer(store)

	srvClosed := make(chan struct{}, 1)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("Server starting, listening on 0.0.0.0:%d...\n", appConfig.port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Error in ListenAndServe(): %v\n", err)
		}
		srvClosed <- struct{}{}
	}()

	shutdown := func(waitForSrv bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Printf("Shutting down...\n")
		if waitForSrv {
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("Error shutting down server: %v\n", err)
			}
		}
		if err := store.Close(ctx); err != nil {
			log.Printf("Error disconnecting from database: %v\n", err)
		}
		stop()
	}

	select {
	case <-srvClosed:
		shutdown(false)
	case <-ctx.Done():
		shutdown(true)
	}

	log.Println("Server has shut down")
}

func config() {
	mongoURI := getEnvString("BOOKLIBRARY_MONGOURI", "mongodb://localhost")
	port := getEnvInt("BOOKLIBRARY_PORT", 8000)
	db := getEnvString("BOOKLIBRARY_DB", "library_database")
	coll := getEnvString("BOOKLIBRARY_COLLECTION", "books")

	flag.IntVar(&(appConfig.port), "port", port, "HTTP port to listen on")
	flag.StringVar(&(appConfig.mongoURI), "mongoURI", mongoURI, "MongoDB URI to connect to")
	flag.StringVar(&(appConfig.db), "db", db, "MongoDB database")
	flag.StringVar(&(appConfig.collection), "collection", coll, "MongoDB collection")
	flag.Parse()
}

func newStorage() (*mongo.MongoCollectionStore, error) {
	log.Printf("Connecting to MongoDB at %q.\n", appConfig.mongoURI)
	s, err := mongo.NewStorage(appConfig.mongoURI, appConfig.db, appConfig.collection)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to MongoDB at %q.\n", appConfig.mongoURI)
	return s, nil
}

func newServer(store booklibrary.Store) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", appConfig.port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      api.NewAPIHandler(store),
	}
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
		log.Printf("Error parsing value %q from %q as int, using default %d: %v\n", envStr, name, value, err)
		return value
	}
	return env
}
