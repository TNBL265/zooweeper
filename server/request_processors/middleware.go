package request_processors

import (
	"bytes"
	"container/heap"
	"encoding/json"
	"github.com/fatih/color"
	"github.com/tnbl265/zooweeper/request_processors/data"
	"github.com/tnbl265/zooweeper/zab"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// QueueMiddleware to order Transaction using PriorityQueue
func (rp *RequestProcessor) QueueMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the timestamp
		timestamp, _ := func(r *http.Request) (string, error) {
			// Read the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return "", err
			}

			// Decode the body into the struct
			var data data.Data
			err = json.Unmarshal(body, &data)
			if err != nil {
				return "", err
			}

			// Close and replace the original body so it can be read again if necessary
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			// Return the timestamp
			return data.Timestamp, nil
		}(r)

		// Use minPriorityQueue to ensure FIFO Client Order
		item := &Transaction{
			Request:   r,
			Timestamp: timestamp,
		}
		heap.Push(&rp.pq, item)
		for rp.pq.Peek() != item {
			time.Sleep(time.Second)
		}
		next.ServeHTTP(w, r)
		heap.Pop(&rp.pq)
	})
}

// WriteOpsMiddleware to establish some form of Total Order for Transaction using PriorityQueue
func (rp *RequestProcessor) WriteOpsMiddleware(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zNode, err := rp.Zab.ZTree.GetLocalMetadata()
		if err != nil {
			log.Println("WriteOpsMiddleware Error:", err)
			return
		}

		if zNode.NodePort != zNode.Leader {
			// Follower will forward Request to Leader
			color.HiBlue("%s forwarding request to leader %s", zNode.NodePort, zNode.Leader)
			resp, err := rp.Zab.ForwardRequestToLeader(r)
			if err != nil {
				// Handle error
				http.Error(w, "Failed to forward request", http.StatusInternalServerError)
				return
			}

			// Copy Header and Status code
			for name, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}
			w.WriteHeader(resp.StatusCode)

			io.Copy(w, resp.Body)
			return
		} else {
			// Leader will Propose, wait for Acknowledge, before Commit
			data := rp.Zab.CreateMetadataFromPayload(w, r)
			for rp.Zab.ProposalState() != zab.COMMITTED {
				// Propose in sequence to ensure Linearization Write
				time.Sleep(time.Second)
			}
			rp.Zab.StartProposal(data)
			return
		}
	})
}
