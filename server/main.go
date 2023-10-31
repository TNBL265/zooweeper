package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	ZNodeHandlers "github.com/tnbl265/zooweeper/database/handlers"
	ensemble "github.com/tnbl265/zooweeper/ensemble"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, _ := strconv.Atoi(portStr)

	leader := 8080
	allServers := []int{8080, 8081, 8082}

	var dbPath string

	switch port {
	case 8080:
		dbPath = "database/zooweeper-metadata-0.db"
	case 8081:
		dbPath = "database/zooweeper-metadata-1.db"
	case 8082:
		dbPath = "database/zooweeper-metadata-2.db"
	default:
		log.Fatalf("Only support ports 8080, 8081 or 8082")
	}

	// Initialization Server
	log.Println("ZooWeeper Server started on port:", port)
	server := ensemble.NewServer(port, leader, allServers)

	// Connect to the Database
	log.Println("Connecting to", dbPath)
	db, err := server.Rp.Zab.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	server.Rp.Zab.ZTree = &ZNodeHandlers.ZTree{DB: db}
	defer func(connection *sql.DB) {
		err := connection.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(server.Rp.Zab.ZTree.Connection())

	// Start Server
	log.Printf("Starting Server (%s) on port %s\n", server.State(), portStr)
	err = http.ListenAndServe(fmt.Sprintf(":"+portStr), server.Rp.Routes())
	if err != nil {
		log.Fatal(err)
	}
}
