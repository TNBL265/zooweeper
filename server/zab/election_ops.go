package zooweeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/tnbl265/zooweeper/database/models"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (ab *AtomicBroadcast) Ping(portStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload models.HealthCheck
		err := ab.readJSON(w, r, &requestPayload)
		if err != nil {
			ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		color.Magenta("Received Healthcheck from Port:%s, Message:%s \n", requestPayload.PortNumber, requestPayload.Message)

		payload := models.HealthCheck{
			Message:    "pong",
			PortNumber: portStr,
		}

		_ = ab.writeJSON(w, http.StatusOK, payload)
	}
}

func (ab *AtomicBroadcast) SelfElectLeaderRequest(portStr string) http.HandlerFunc {
	const REQUEST_TIMEOUT = 10 // Arbitrary wait timer to simulate response time arrival
	return func(w http.ResponseWriter, r *http.Request) {
		hasFailedElection := false

		var requestPayload models.ElectLeaderRequest
		err := ab.readJSON(w, r, &requestPayload)
		if err != nil {
			ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		color.Magenta("Received election message from Port:%s \n", requestPayload.IncomingPort)

		incomingPortNumber, _ := strconv.Atoi(requestPayload.IncomingPort)
		currentPortNumber, _ := strconv.Atoi(portStr)

		payload := models.ElectLeaderResponse{
			IsSuccess: strconv.FormatBool(incomingPortNumber > currentPortNumber),
		}
		metadata, _ := ab.ZTree.GetLocalMetadata()
		allServers := strings.Split(metadata.Servers, ",")

		// If it has a better node number than the incoming one, send a value updwards to all nodes higher than it.
		if incomingPortNumber <= currentPortNumber {
			// Send self elect message to all nodes that is higher than current node
			for _, outgoingPort := range allServers {
				outgoingPortNumber, _ := strconv.Atoi(outgoingPort)
				if outgoingPortNumber < currentPortNumber || outgoingPortNumber == currentPortNumber {
					continue
				}

				// make a http request
				client := &http.Client{}
				portURL := fmt.Sprintf("%s", outgoingPort)

				url := fmt.Sprintf(ab.BaseURL + ":" + portURL + "/electLeader")
				var electMessage models.ElectLeaderRequest = models.ElectLeaderRequest{
					IncomingPort: fmt.Sprintf("%d", currentPortNumber),
				}
				jsonData, _ := json.Marshal(electMessage)

				req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
				if err != nil {
					log.Println(err)
					continue
				}
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")

				ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT*time.Second)
				defer cancel()

				req = req.WithContext(ctx)
				// CONNECTION
				resp, err := client.Do(req)
				if err != nil || resp == nil {
					log.Println("Timeout issue!")
					log.Println(err)
					continue
				} else if err != nil {
					log.Println(err)
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

				// If higher node response, determine if election should fail.
				responseBool, _ := strconv.ParseBool(responseObject.IsSuccess)
				if !responseBool {
					hasFailedElection = true
				}

			}

		}
		if !hasFailedElection {
			color.Green("Port %s has won election", portStr)
		} else {
			color.Red("Port %s has lost election", portStr)
		}

		// Declare itself leader to all other nodes if node succeeds
		if !hasFailedElection {
			ab.declareLeaderRequest(portStr, allServers)
		}
		_ = ab.writeJSON(w, http.StatusOK, payload)
	}
}

// Send request to all other nodes that outgoing port is a leader.
func (ab *AtomicBroadcast) declareLeaderRequest(portStr string, allServers []string) {
	for _, outgoingPort := range allServers {
		//make a request
		client := &http.Client{}
		portURL := fmt.Sprintf("%s", outgoingPort)

		url := fmt.Sprintf(ab.BaseURL + ":" + portURL + "/declareLeaderReceive")
		var electMessage = models.DeclareLeaderRequest{
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

// DeclareLeaderReceive send response to all other nodes that incoming port is a leader.
func (ab *AtomicBroadcast) DeclareLeaderReceive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//send information to all servers
		var requestPayload models.DeclareLeaderRequest
		err := ab.readJSON(w, r, &requestPayload)
		if err != nil {
			ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}

		color.Cyan("%s", requestPayload.IncomingPort)
		ab.ZTree.UpdateFirstLeader(requestPayload.IncomingPort)
	}
}
