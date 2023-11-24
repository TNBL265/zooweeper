package main

import (
	"bytes"
	"context"
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
	zooweeper "github.com/tnbl265/zooweeper/zab"
)

func main() {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}
	port, _ := strconv.Atoi(portStr)

	var state ensemble.ServerState
	leader := 8082
	allServers := []int{8080, 8081, 8082, 8083, 8084, 8085, 8086, 8087}

	var dbPath string
	switch port {
	case 8080:
		dbPath = "database/zooweeper-metadata-0.db"
		state = ensemble.FOLLOWING
	case 8081:
		dbPath = "database/zooweeper-metadata-1.db"
		state = ensemble.FOLLOWING
	case 8082:
		dbPath = "database/zooweeper-metadata-2.db"
		state = ensemble.FOLLOWING
	case 8083:
		dbPath = "database/zooweeper-metadata-3.db"
		state = ensemble.FOLLOWING
	case 8084:
		dbPath = "database/zooweeper-metadata-4.db"
		state = ensemble.FOLLOWING
	case 8085:
		dbPath = "database/zooweeper-metadata-5.db"
		state = ensemble.FOLLOWING
	case 8086:
		dbPath = "database/zooweeper-metadata-6.db"
		state = ensemble.FOLLOWING
	case 8087:
		dbPath = "database/zooweeper-metadata-7.db"
		state = ensemble.LEADING
	default:
		log.Fatalf("Only support ports 8080 to 8087")
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

	// Listen for errors
	var err error
	myAB := &AtomicBroadcastCopy{AtomicBroadcast: server.Rp.Zab}
	go myAB.wakeupLeaderElection(server, port, leader, allServers)
	go myAB.listenForLeaderElection(server, port, leader, allServers)
	go func() {
		_, err := ping(server, portStr)
		if err != nil {
			// do sth
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

// This function pings all other servers every arbitruary time
// If the ping response takes more than an arbitruary time, or the connection is refused from the other server,
// then call the Error Channel with error information.
func ping(server *ensemble.Server, currentPort string) (string, error) {
	const PING_TIMEOUT = 5    // Arbitruary wait timer to simulate response time arrival
	const REQUEST_TIMEOUT = 2 // Arbitruary wait timer to simulate response time arrival
	for {

		time.Sleep(time.Second * time.Duration(PING_TIMEOUT))
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
				continue
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")

			color.Blue("Sending Ping to Port: %s", v)

			ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT*time.Second)
			defer cancel()

			req = req.WithContext(ctx)

			// CONNECTION
			resp, err := client.Do(req)
			if resp == nil && err != nil {
				if ctxErr := ctx.Err(); ctxErr == context.DeadlineExceeded {
					color.Red("Timeout Occurred!")
				}
			}
			if err != nil || resp == nil {
				color.Red("Error sending ping:")
				log.Println(err)

				errorData := models.HealthCheckError{
					Error:     err,
					ErrorPort: v,
					IsWakeup:  false,
				}
				server.Rp.Zab.ErrorLeaderChan <- errorData
				continue
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

type AtomicBroadcastCopy struct {
	zooweeper.AtomicBroadcast // Embedding the type from the external package
}

func (ab *AtomicBroadcastCopy) wakeupLeaderElection(server *ensemble.Server, port int, leader int, allServers []int) {
	for {
		select {
		case <-time.After(time.Second * 2):
			// On wake-up, start leader-election
			data := models.HealthCheckError{
				Error:     nil,
				ErrorPort: fmt.Sprintf("%d", port),
				IsWakeup:  true,
			}
			ab.ErrorLeaderChan <- data
			return
		}
	}
}
func (ab *AtomicBroadcastCopy) listenForLeaderElection(server *ensemble.Server, port int, leader int, allServers []int) {
	for {
		select {
		case errorData := <-ab.ErrorLeaderChan:
			errorPortNumber, _ := strconv.Atoi(errorData.ErrorPort)
			if errorPortNumber == leader || errorData.IsWakeup {
				if errorPortNumber == leader {
					color.Magenta("Error from port %d healthcheck! Starting leader election here...", errorPortNumber)
				} else if errorData.IsWakeup {
					color.Magenta("Port %d just woke up! Starting leader election here...", port)
				}

				ab.startLeaderElection(server, port, allServers)
			}
		}
	}
}

// Starts the leader election by calling its own server handler with information of the port information.
func (ab *AtomicBroadcastCopy) startLeaderElection(server *ensemble.Server, currentPort int, allServers []int) {

	client := &http.Client{}
	portURL := fmt.Sprintf("%d", currentPort)

	url := fmt.Sprintf(ab.BaseURL + ":" + portURL + "/electLeader")
	var electMessage models.ElectLeaderRequest = models.ElectLeaderRequest{
		IncomingPort: fmt.Sprintf("%d", currentPort),
	}
	jsonData, _ := json.Marshal(electMessage)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
}
