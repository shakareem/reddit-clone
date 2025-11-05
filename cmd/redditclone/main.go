package main

import (
	"log"
	"redditclone/internal/server"
)

func main() {
	server := server.NewServer()
	err := server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
