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

	startPortStr := os.Getenv("START_PORT")
	if startPortStr == "" {
		startPortStr = "8080"
	}
	startPort, _ := strconv.Atoi(startPortStr)

	// Default local testing with 3 ZooWeeper servers
	endPortStr := os.Getenv("END_PORT")
	if endPortStr == "" {
		endPortStr = "8082"
	}
	endPort, _ := strconv.Atoi(endPortStr)

	leader := endPort
	allServers := make([]int, 0, endPort-startPort+1)
	for p := startPort; p <= endPort; p++ {
		allServers = append(allServers, p)
	}

	var dbPath string
	if port >= startPort && port <= endPort {
		dbPath = fmt.Sprintf("ztree/zooweeper-metadata-%d.db", port-startPort)
	} else {
		log.Fatalf("Only support ports %d to %d", startPort, endPort)
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
