package gopaste

import (
	"flag"
	"fmt"
	"os"
)

const (
	DefaultDriver   = "sqlite3"
	DefaultDatabase = "gopaste.sqlite"
	DefaultPort     = 80
)

type Config struct {
	DbDriver     string
	DbSource     string
	Port         uint
	ExternalHost string
	HubotHost    string
}

// ParseConfig creates a new Config object by reading the command-line arguments.
func ParseConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.DbDriver, "db-driver", DefaultDriver, "Database driver")
	flag.StringVar(&config.DbSource, "db-source", DefaultDatabase, "Database source")
	flag.UintVar(&config.Port, "port", DefaultPort, "HTTP server port")
	flag.StringVar(&config.ExternalHost, "external-host", "", "Gopaste hostname for external links")
	flag.StringVar(&config.HubotHost, "hubot-host", "", "Hubot location")
	flag.Parse()

	if config.ExternalHost == "" {
		localhost, err := os.Hostname()
		if err != nil {
			localhost = "localhost"
		}
		config.ExternalHost = fmt.Sprintf("%s:%d", localhost, config.Port)
	}

	return config
}
