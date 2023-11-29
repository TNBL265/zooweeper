package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/tnbl265/zooweeper/ensemble"
	"github.com/tnbl265/zooweeper/request_processors/data"

	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tnbl265/zooweeper/ztree"
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

	// Regular Health Checks and start Leader Election if need to
	var err error
	go server.Rp.Zab.WakeupLeaderElection(port)
	go server.Rp.Zab.ListenForLeaderElection(port, leader)
	go func() {
		_, err := ping(server)
		if err == nil {
			fmt.Println("Ping successful")
		}
	}()

	initZNode(server, port, leader, allServers)

	err = http.ListenAndServe(fmt.Sprintf(":"+portStr), server.Rp.Routes(portStr))
	if err != nil {
		log.Fatal(err)
	}

}

// initZNode insert first Znode to self-identify
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
	server.Rp.Zab.ZTree.InsertMetadataWithParentId(metadata, 0)
}

// ping all other servers
// If the ping response takes more than an arbitrary time, or the connection is refused from the other server,
// then call the Error Channel with error information to signal Leader Election
func ping(server *ensemble.Server) (string, error) {
	const PING_TIMEOUT = 5
	const REQUEST_TIMEOUT = 2

	for {
		time.Sleep(time.Second * time.Duration(PING_TIMEOUT))
		zNode, _ := server.Rp.Zab.ZTree.GetLocalMetadata()
		currentPort := zNode.NodePort
		//startTime := time.Now()

		var healthCheck = data.HealthCheck{
			Message:    "ping!",
			PortNumber: currentPort,
		}
		jsonData, _ := json.Marshal(healthCheck)

		servers := strings.Split(zNode.Servers, ",")
		for _, otherPort := range servers {
			if currentPort == otherPort {
				continue
			}
			client := &http.Client{}
			url := fmt.Sprintf(server.Rp.Zab.BaseURL + ":" + otherPort + "/")

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Println(err)
				continue
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("X-Sender-Port", zNode.NodePort)

			color.Green("Ping %s", otherPort)

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
				color.Red("Error sending ping to %s", otherPort)

				errorData := data.HealthCheckError{
					Error:     err,
					ErrorPort: otherPort,
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
			var responseObject data.HealthCheck
			err = json.Unmarshal(resBody, &responseObject)
			if err != nil {
				log.Println(err)
				continue
			}
			// Port number
			color.Green("%s Pong", responseObject.PortNumber)

			// Elapsed Time
			//endTime := time.Now()
			//elapsedTime := endTime.Sub(startTime)
			//color.Green("Time taken to get this Pong return: %s\n", elapsedTime)
		}
	}
}
