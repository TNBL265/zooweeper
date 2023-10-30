package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"

	ZNodeHandlers "github.com/tnbl265/zooweeper/database/handlers"
	requestProcessor "github.com/tnbl265/zooweeper/request_processors"
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
	var rp requestProcessor.RequestProcessor

	// Connect to the Database
	log.Println("Connecting to", dbPath)
	db, err := rp.Zab.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	rp.Zab.ZTree = &ZNodeHandlers.ZTree{DB: db}
	//close when it is done
	defer func(connection *sql.DB) {
		err := connection.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rp.Zab.ZTree.Connection())

	// Start a Web Server
	log.Println("Starting rplication on port", port)
	err = http.ListenAndServe(fmt.Sprintf(":"+port), rp.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
