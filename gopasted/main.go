package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	flag "github.com/ogier/pflag"
	"github.com/wisnij/gopaste"
	"log"
)

const (
	DefaultDatabase = "gopaste.sqlite"
	DefaultPort     = 80
)

func main() {
	dbFile := flag.String("source", DefaultDatabase, "Database data source")
	port := flag.Uint("port", DefaultPort, "HTTP server port")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	dbh, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		log.Fatalf("Error opening %s: %v\n", *dbFile, err)
	}
	defer dbh.Close()

	server, err := gopaste.New(dbh)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = server.ListenAndServe(addr)
	if err != nil {
		log.Fatal(err.Error())
	}
}
