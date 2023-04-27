package main

import (
	"goDrive/internal"
	"log"
	"net/http"
	"goDrive/internal/service"
)

const port = ":7789"

func main() {
	service.InitProps()

	service.InitPgPool()
	defer service.CloseDb()

	r := internal.SetupRoutes()

	log.Printf("Server started on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))	
}
