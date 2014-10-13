package gopaste

import (
	"flag"
)

const (
	DefaultDriver   = "sqlite3"
	DefaultDatabase = "gopaste.sqlite"
	DefaultPort     = 80
)

type Config struct {
	DbDriver string
	DbSource string
	Port     uint
}

// ParseConfig creates a new Config object by reading the command-line arguments.
func ParseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.DbDriver, "db-driver", DefaultDriver, "Database driver")
	flag.StringVar(&config.DbSource, "db-source", DefaultDatabase, "Database source")
	flag.UintVar(&config.Port, "port", DefaultPort, "HTTP server port")

	flag.Parse()
	return config
}
