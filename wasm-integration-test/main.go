package main

import (
	"log"
	"net/http"
)

func main() {

	err := http.ListenAndServe(":9090", http.FileServer(http.Dir("./assets")))
	if err != nil {
		log.Fatal("Failed to start server", err)
	}

}
