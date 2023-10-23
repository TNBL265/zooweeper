package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"

	api "github.com/tnbl265/zooweeper/server/api"
)

const port = 8080

func main() {
	// Initialization code here
	fmt.Println("ZooWeeper Server started.")

	// Set Application Config
	var app api.Application

	// Connect to the Database
	log.Println("Connecting to sqlite3 database")
	db, err := sql.Open("sqlite3", "zooweeper-database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Start a Web Server
	log.Println("Starting application on port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
