package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	api "github.com/tnbl265/zooweeper/api"
	dbrepo "github.com/tnbl265/zooweeper/database/dbrepo"
)

const port = 8080

func main() {
	// Initialization code here
	fmt.Println("ZooWeeper Server started.")

	// Set Application Config
	var app api.Application

	// Connect to the Database
	log.Println("Connecting to sqlite3 database")
	db, err := app.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = &dbrepo.SQLiteDBRepo{DB: db}
	//close when it is done
	defer app.DB.Connection().Close()

	// Start a Web Server
	log.Println("Starting application on port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
