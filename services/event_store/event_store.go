package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbname := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)

	if err != nil {
		log.Fatalln("Couldn't connect to a DB")
	}

	defer db.Close()

	// check db
	err = db.Ping()
	if err != nil {
		log.Fatalln("DB does not respond")
	}
	fmt.Println("Connected!")

}
