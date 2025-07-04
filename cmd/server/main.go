package main

import (
	"log"
	"task/internal/tcp"
)

func main() {
	srv, err := tcp.NewServer(":9000")
	if err != nil {
		log.Fatal(err)
	}

	err = srv.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
