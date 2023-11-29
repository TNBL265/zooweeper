package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tnbl265/zooweeper/ensemble"
	"github.com/tnbl265/zooweeper/ztree"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, _ := strconv.Atoi(portStr)

	leader := 8082
	allServers := []int{8080, 8081, 8082} //, 8083, 8084, 8085, 8086, 8087}

	var dbPath string
	if port >= 8080 && port <= 8087 {
		dbPath = fmt.Sprintf("ztree/zooweeper-metadata-%d.db", port-8080)
	} else {
		log.Fatalf("Only support ports 8080 to 8087")
	}

	// Start Server
	server := ensemble.NewServer(dbPath)
	log.Printf("Starting Server on port %s\n", portStr)
	defer func(connection *sql.DB) {
		err := connection.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(server.Rp.Zab.ZTree.Connection())

	// Regular Health Checks and start Leader Election once failure detected
	var err error
	go server.Rp.Zab.WakeupLeaderElection(port)
	go server.Rp.Zab.ListenForLeaderElection(port, leader)
	go server.Rp.Zab.StartHealthCheck()

	initZNode(server, port, leader, allServers)

	err = http.ListenAndServe(fmt.Sprintf(":"+portStr), server.Rp.Routes(portStr))
	if err != nil {
		log.Fatal(err)
	}

}

// initZNode insert first ZNode to self-identify in the current ZooWeeper ensemble
func initZNode(server *ensemble.Server, port, leader int, allServers []int) {
	existFirstNode, _ := server.Rp.Zab.ZTree.ZNodeIdExists(1)
	if existFirstNode {
		return
	}
	var result []string
	for _, server := range allServers {
		result = append(result, strconv.Itoa(server))
	}

	allServersStr := strings.Join(result, ",")

	metadata := ztree.Metadata{
		NodePort: strconv.Itoa(port),
		Leader:   strconv.Itoa(leader),
		Servers:  allServersStr,
	}
	server.Rp.Zab.ZTree.InsertFirstMetadata(metadata)
}
