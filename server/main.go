package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tnbl265/zooweeper/database/models"
	ensemble "github.com/tnbl265/zooweeper/ensemble"
)

func main() {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, _ := strconv.Atoi(portStr)

	var state ensemble.ServerState
	leader := 8080
	allServers := []int{8080, 8081, 8082}

	var dbPath string
	switch port {
	case 8080:
		dbPath = "database/zooweeper-metadata-0.db"
		state = ensemble.LEADING
	case 8081:
		dbPath = "database/zooweeper-metadata-1.db"
		state = ensemble.FOLLOWING
	case 8082:
		dbPath = "database/zooweeper-metadata-2.db"
		state = ensemble.FOLLOWING
	default:
		log.Fatalf("Only support ports 8080, 8081 or 8082")
	}

	// Start Server
	server := ensemble.NewServer(port, leader, state, allServers, dbPath)
	log.Printf("Starting Server (%s) on port %s\n", server.State(), portStr)
	defer func(connection *sql.DB) {
		err := connection.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(server.Rp.Zab.ZTree.Connection())

	var err error

	go func() {
		err = ping(server, portStr)
		if err != nil {
			fmt.Println("Ping error:", err)
		} else {
			fmt.Println("Ping successful")
		}
	}()
	// I
	initZNode(server, port, leader, allServers)

	err = http.ListenAndServe(fmt.Sprintf(":"+portStr), server.Rp.Routes(portStr))
	if err != nil {
		log.Fatal(err)
	}
}

func ping(server *ensemble.Server, currentPort string) error {
	for {
		time.Sleep(time.Second * time.Duration(2))
		// start timer here
		startTime := time.Now()

		var healthCheck models.HealthCheck = models.HealthCheck{
			Message:    "ping!",
			PortNumber: currentPort,
		}
		jsonData, _ := json.Marshal(healthCheck)

		metadata, _ := server.Rp.Zab.ZTree.GetLocalMetadata()
		servers := strings.Split(metadata.Servers, ",")
		for _, v := range servers {
			if currentPort == v {
				continue
			}
			client := &http.Client{}
			url := fmt.Sprintf(server.Rp.Zab.BaseURL + ":" + v + "/")

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Println(err)
				return err
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")

			color.Blue("Sending Ping to Port: %s", v)

			// CONNECTION
			resp, err := client.Do(req)
			time.Sleep(time.Second * time.Duration(2)) // Arbitruary wait timer to simulate response time.
			if err != nil {
				color.Red("Error sending ping:")
				log.Println(err)
				continue // TODO: server is down, to perform leader election.
			}

			// REPLY
			defer resp.Body.Close()
			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				continue
			}
			var responseObject models.HealthCheck
			err = json.Unmarshal(resBody, &responseObject)
			if err != nil {
				log.Println(err)
				continue
			}

			// Status Code of Pong
			statusCode := resp.StatusCode
			color.Green("Status Code received: %d\n", statusCode)

			// Port number
			color.Green("Pong return from Server Port Number: %s \n", responseObject.PortNumber)

			// Elapsed Time
			endTime := time.Now()
			elapsedTime := endTime.Sub(startTime)
			color.Green("Time taken to get this Pong return: %s\n", elapsedTime)
		}
		fmt.Printf("===================\n")
	}
}

// initZNode insert first Znode to self-identify
func initZNode(server *ensemble.Server, port, leader int, allServers []int) {
	existFirstNode, _ := server.Rp.Zab.ZTree.NodeIdExists(1)
	if existFirstNode {
		return
	}
	var result []string
	for _, server := range allServers {
		result = append(result, strconv.Itoa(server))
	}

	allServersStr := strings.Join(result, ",")

	metadata := models.Metadata{
		NodeIp:  strconv.Itoa(port),
		Leader:  strconv.Itoa(leader),
		Servers: allServersStr,
	}
	server.Rp.Zab.ZTree.InsertMetadata(metadata, 0)
}
