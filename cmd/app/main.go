package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	db := repo.ConnectDB()
	defer db.close()

	logger, logfile := utils.NewLogger()
	defer logfile.close()

	baseHandler := handler.NewBaseHandler()
	repos := repo.New()
	services := service.New()
	handlers := handler.New()

	mux := api.Router()
	log.Fatalln(http.ListenAndServe(":8080"), mux)
}
