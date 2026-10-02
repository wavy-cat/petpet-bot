package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", discordInteractions)
	log.Printf("HTTP server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
