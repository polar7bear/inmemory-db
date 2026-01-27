package main

import (
	"inmemory-db/internal/server"
	"log"
)

func main() {
	server := server.New(":6379")

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	server.Start()
}
