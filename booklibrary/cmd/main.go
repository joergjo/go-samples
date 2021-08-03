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

	go func() {
		log.Printf("Server starting, listening on 0.0.0.0:%d...\n", appConfig.port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Fatal error in ListenAndServe(): %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Fatal error shutting down server: %v\n", err)
	}
	log.Println("Server has shut down")
}

func config() {
	mongoURI := os.Getenv("BOOKLIBRARY_MONGOURI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost"
	}
	port, err := strconv.Atoi(os.Getenv(("BOOKLIBRARY_PORT")))
	if err != nil {
		port = 5000
	}
	db := os.Getenv("BOOKLIBRARY_DB")
	if db == "" {
		db = "library_database"
	}
	coll := os.Getenv("BOOKLIBRARY_COLLECTION")
	if coll == "" {
		coll = "books"
	}
	flag.IntVar(&(appConfig.port), "port", port, "Port number to listen on")
	flag.StringVar(&(appConfig.mongoURI), "mongoURI", mongoURI, "MongoDB URI to connect to")
	flag.StringVar(&(appConfig.db), "db", db, "Name of MongoDB database")
	flag.StringVar(&(appConfig.collection), "collection", coll, "Name of MongoDB collection")
	flag.Parse()
}

func newStorage() (booklibrary.Store, error) {
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
