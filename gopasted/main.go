package main

import (
	"github.com/wisnij/gopaste"
	"log"
)

func main() {
	config := gopaste.ParseConfig()
	server, err := gopaste.New(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
