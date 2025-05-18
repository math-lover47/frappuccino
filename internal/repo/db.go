package repo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var (
	host        = os.Getenv("DB_HOST")
	user        = os.Getenv("DB_USER")
	password    = os.Getenv("DB_PASSWORD")
	name        = os.Getenv("DB_NAME")
	port, exist = os.LookupEnv("DB_PORT")
)

func ConnectDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping failed, not ponged: %v", err)
	}

	log.Println("Successfully connect to db!")
	return db
}
