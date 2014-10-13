package gopaste

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Config   *Config
	Database *sql.DB
}

// New creates a new Gopaste server object and opens its database connection.
func New(config *Config) (*Server, error) {
	server := &Server{Config: config}

	err := server.initDb()
	if err != nil {
		return nil, err
	}

	return server, nil
}

// ListenAndServe starts the server listening for incoming requests on the
// specified port.
func (s *Server) ListenAndServe() error {
	// use ServeMux to get path cleaning, etc. for free
	mux := http.NewServeMux()
	mux.Handle("/", s)

	addr := fmt.Sprintf(":%d", s.Config.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("[server] listening on %s", addr)
	err := httpServer.ListenAndServe()
	if err != nil {
		return err
	}

	log.Print("[server] exiting")
	return nil
}
