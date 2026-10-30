package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Imanghvs/crudgo/internal/adapter/httpapi"
)

const port = 8080

func main() {
	log.Printf("Starting the server on port %d", port)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", httpapi.Health)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatal(err)
	}
}
