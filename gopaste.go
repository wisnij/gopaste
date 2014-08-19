package gopaste

import (
	"database/sql"
	"log"
	"net/http"
)

type Server struct {
	Database *sql.DB
}

func New(dbh *sql.DB) (*Server, error) {
	err := initDb(dbh)
	if err != nil {
		return nil, err
	}

	server := &Server{Database: dbh}
	return server, nil
}

func (s *Server) ListenAndServe(addr string) error {
	// use ServeMux to get path cleaning, etc. for free
	mux := http.NewServeMux()
	mux.Handle("/", s)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("[server] listening on %s", addr)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	log.Print("[server] exiting")
	return nil
}
