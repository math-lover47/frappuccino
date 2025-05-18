package main

import (
	"frappuccino/internal/api"
	"frappuccino/internal/api/handler"
	"frappuccino/internal/repo"
	"frappuccino/internal/service"
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

	baseHandler := handler.NewBaseHandler(*logger)
	repos := repo.New(db)
	services := service.New(repos)
	handlers := handler.New(services, baseHandler)

	mux := api.Router(handlers)
	log.Fatalln(http.ListenAndServe(":8080"), mux)
}
