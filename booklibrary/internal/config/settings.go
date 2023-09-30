package config

type Settings struct {
	Port       int
	MongoURI   string
	Db         string
	Collection string
	Debug      bool
}
