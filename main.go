package main

import (
	"log"
	"net/http"

	"github.com/Brightwater/goDrive/internal"
	"github.com/Brightwater/goDrive/internal/cron"
	"github.com/Brightwater/goDrive/internal/db"
	"github.com/Brightwater/goDrive/internal/service"
)

func main() {

	service.InitProps()
	port := ":" + service.AppConfig.HttpPort

	db.InitPgPool()
	defer db.CloseDb()

	cron.StartCronService()
	defer cron.StopCronService()

	r := internal.SetupRoutes()

	log.Printf("Server started on port %s", port)

	log.Fatal(http.ListenAndServe(port, r))
}
