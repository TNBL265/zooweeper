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

type ElectionOps struct {
	ab *AtomicBroadcast
}

func (eo *ElectionOps) Ping(portStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload models.HealthCheck
		err := eo.ab.readJSON(w, r, &requestPayload)
		if err != nil {
			eo.ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}

		payload := models.HealthCheck{
			Message:    "pong",
			PortNumber: portStr,
		}

		_ = eo.ab.writeJSON(w, http.StatusOK, payload)
	}
}

func (eo *ElectionOps) SelfElectLeaderRequest(portStr string) http.HandlerFunc {
	const REQUEST_TIMEOUT = 10 // Arbitrary wait timer to simulate response time arrival
	return func(w http.ResponseWriter, r *http.Request) {
		hasFailedElection := false

		var requestPayload models.ElectLeaderRequest
		err := eo.ab.readJSON(w, r, &requestPayload)
		if err != nil {
			eo.ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}

		incomingPortNumber, _ := strconv.Atoi(requestPayload.IncomingPort)
		currentPortNumber, _ := strconv.Atoi(portStr)

		payload := models.ElectLeaderResponse{
			IsSuccess: strconv.FormatBool(incomingPortNumber > currentPortNumber),
		}
		metadata, _ := eo.ab.ZTree.GetLocalMetadata()
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

				url := fmt.Sprintf(eo.ab.BaseURL + ":" + portURL + "/electLeader")
				var electMessage models.ElectLeaderRequest = models.ElectLeaderRequest{
					IncomingPort: fmt.Sprintf("%d", currentPortNumber),
				}
				jsonData, _ := json.Marshal(electMessage)

				color.Cyan("%s sending Self-Elect to %s", portStr, outgoingPort)
				req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")

				ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT*time.Second)
				defer cancel()

				req = req.WithContext(ctx)
				// CONNECTION
				resp, err := client.Do(req)
				if err != nil || resp == nil {
					color.Red("Timeout from %s", outgoingPort)
					continue
				}

				defer resp.Body.Close()
				resBody, _ := ioutil.ReadAll(resp.Body)

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
			color.Cyan("%s won election", portStr)
		} else {
			color.Cyan("%s lost election", portStr)
		}

		// Declare itself leader to all other nodes if node succeeds
		if !hasFailedElection {
			eo.ab.declareLeaderRequest(portStr, allServers)
			eo.ab.syncMetadata()
		}
		_ = eo.ab.writeJSON(w, http.StatusOK, payload)
	}
}

// DeclareLeaderReceive send response to all other nodes that incoming port is a leader.
func (eo *ElectionOps) DeclareLeaderReceive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload models.DeclareLeaderRequest
		err := eo.ab.readJSON(w, r, &requestPayload)
		if err != nil {
			eo.ab.errorJSON(w, err, http.StatusBadRequest)
			return
		}
		eo.ab.ZTree.UpdateFirstLeader(requestPayload.IncomingPort)
	}
}
