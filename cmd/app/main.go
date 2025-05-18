package main

import (
	"frappuccino/internal/api"
	"frappuccino/internal/api/handlers"
	"frappuccino/internal/repo"
	"frappuccino/internal/services"
	"frappuccino/utils"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	db := repo.ConnectDB()
	defer db.Close()

	logger, logfile := utils.NewLogger()
	defer logfile.Close()

	baseHandler := handlers.NewBaseHandler(logger)
	repos := repo.New(db)
	services := services.New(repos)
	handlers := handlers.New(services, baseHandler)

	mux := api.Router(handlers)
	log.Fatalln(http.ListenAndServe(":8080", mux))
}
