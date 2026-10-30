package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = 8080

func main() {
	log.Printf("Starting the server on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
