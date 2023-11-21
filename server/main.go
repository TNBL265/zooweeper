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
	myAB := &AtomicBroadcastCopy{AtomicBroadcast: server.Rp.Zab}
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

type AtomicBroadcastCopy struct {
	zooweeper.AtomicBroadcast // Embedding the type from the external package
}

func (ab *AtomicBroadcastCopy) listenForLeaderElection(server *ensemble.Server, port int, leader int, allServers []int) {
	for {
		select {
		case errorData := <-ab.ErrorLeaderChan:
			errorPortNumber, _ := strconv.Atoi(errorData.ErrorPort)
			if errorPortNumber == leader {
				color.Magenta("Error from ping healthcheck! Starting leader election here...")
				startLeaderElection(server, port, allServers)
			}

		}
	}

}

func startLeaderElection(server *ensemble.Server, currentPort int, allServers []int) {
	hasFailedElection := false
	for _, outgoingPort := range allServers {

		if outgoingPort < currentPort || outgoingPort == currentPort {
			continue
		}

		//make a request
		client := &http.Client{}
		portURL := fmt.Sprintf("%d", outgoingPort)

		url := fmt.Sprintf(server.Rp.Zab.BaseURL + ":" + portURL + "/electLeader")
		var electMessage models.ElectLeaderRequest = models.ElectLeaderRequest{
			IncomingPort: fmt.Sprintf("%d", currentPort),
		}
		jsonData, _ := json.Marshal(electMessage)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println(err)
			continue
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		defer resp.Body.Close()
		resBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}
		var responseObject models.ElectLeaderResponse
		err = json.Unmarshal(resBody, &responseObject)
		if err != nil {
			log.Println(err)
			continue
		}

		responseBool, _ := strconv.ParseBool(responseObject.IsSuccess)
		if !responseBool {
			hasFailedElection = true
		}

	}
	if !hasFailedElection {
		declareLeaderRequest(server, fmt.Sprintf("%d", currentPort), allServers)
	}
	color.Red("results is %t", hasFailedElection)
}

func declareLeaderRequest(server *ensemble.Server, portStr string, allServers []int) {
	for _, outgoingPort := range allServers {
		//make a request
		client := &http.Client{}
		portURL := fmt.Sprintf("%d", outgoingPort)

		url := fmt.Sprintf(server.Rp.Zab.BaseURL + ":" + portURL + "/declareLeaderReceive")
		var electMessage models.DeclareLeaderRequest = models.DeclareLeaderRequest{
			IncomingPort: portStr,
		}
		jsonData, _ := json.Marshal(electMessage)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println(err)
			continue
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
	}
}

func ping(server *ensemble.Server, currentPort string) (string, error) {
	const TIMEOUT = 2 // Arbitruary wait timer to simulate response time arrival
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
				continue
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")

			color.Blue("Sending Ping to Port: %s", v)

			ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT*time.Second)
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
