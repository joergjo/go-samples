package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	webapi "github.com/joergjo/go-samples/booklibrary/http"
	"github.com/joergjo/go-samples/booklibrary/mock"
	"github.com/joergjo/go-samples/booklibrary/mongo"
)

var appConfig = struct {
	port       int
	mongoURI   string
	db         string
	collection string
}{}

func init() {
	config()
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[booklibrary] ")
	store, err := newStorage()
	if err != nil {
		log.Fatalf("Failed to create storage implementation: %s", err)
	}
	webAPI := webapi.NewServer(store)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", appConfig.port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      webAPI,
	}
	log.Printf("Server listening on 0.0.0.0:%d...\n", appConfig.port)
	log.Fatal(srv.ListenAndServe())
}

func newStorage() (booklibrary.Storage, error) {
	if appConfig.mongoURI == "" {
		log.Println("Using in-memory data store.")
		s, err := mock.NewStorage(mock.SampleData())
		return s, err
	}
	log.Printf("Connecting to MongoDB at '%s'.\n", appConfig.mongoURI)
	s, err := mongo.NewStorage(appConfig.mongoURI, appConfig.db, appConfig.collection)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to MongoDB at '%s'.\n", appConfig.mongoURI)
	return s, nil
}

func config() {
	mongoURI := os.Getenv("BOOKLIBRARY_MONGOURI")
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
}
