package config

// Settings represents the application settings.
type Settings struct {
	// Port is the port the HTTP server listens on.
	Port int
	// MongoURI is the MongoDB connection string.
	MongoURI string
	// MongoDB is the MongoDB database name.
	Db string
	// Collection is the MongoDB collection name.
	Collection string
	// Debug is the debug mode (verbose logging).
	Debug bool
}
