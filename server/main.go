package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"

	api "github.com/tnbl265/zooweeper/api"
	dbrepo "github.com/tnbl265/zooweeper/database/dbrepo"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var dbPath string

	switch port {
	case "8080":
		dbPath = "database/zooweeper-metadata-0.db"
	case "8081":
		dbPath = "database/zooweeper-metadata-1.db"
	case "8082":
		dbPath = "database/zooweeper-metadata-2.db"
	default:
		log.Fatalf("Unsupported port: %s", port)
	}

	// Initialization code here
	fmt.Println("ZooWeeper Server started on port:", port)

	// Set Application Config
	var app api.Application

	// Connect to the Database
	log.Println("Connecting to", dbPath)
	db, err := app.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	app.DB = &dbrepo.SQLiteDBRepo{DB: db}
	//close when it is done
	defer app.DB.Connection().Close()

	// Start a Web Server
	log.Println("Starting application on port", port)
	err = http.ListenAndServe(fmt.Sprintf(":"+port), app.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
