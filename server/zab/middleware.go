package zooweeper

import (
	"bytes"
	"container/heap"
	"encoding/json"
	"github.com/tnbl265/zooweeper/database/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type RequestItem struct {
	Request   *http.Request
	Timestamp string
}

type PriorityQueue []*RequestItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Timestamp < pq[j].Timestamp
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*RequestItem)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Peek() *RequestItem {
	if len(*pq) == 0 {
		return nil
	}
	return (*pq)[0]
}

func (ab *AtomicBroadcast) QueueMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the timestamp
		timestamp, _ := func(r *http.Request) (string, error) {
			// Read the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return "", err
			}

			// Decode the body into the struct
			var data models.Data
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
		item := &RequestItem{
			Request:   r,
			Timestamp: timestamp,
		}
		heap.Push(&ab.pq, item)
		for ab.pq.Peek() != item {
			time.Sleep(time.Second)
		}
		next.ServeHTTP(w, r)
		heap.Pop(&ab.pq)
	})
}

func (wo *WriteOps) WriteOpsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zNode, err := wo.ab.ZTree.GetLocalMetadata()
		if err != nil {
			log.Println("WriteOpsMiddleware Error:", err)
			return
		}

		if zNode.NodeIp != zNode.Leader {
			// Follower will forward Request to Leader
			log.Println("Forwarding request to leader")
			resp, err := wo.ab.forwardRequestToLeader(r)
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
			data := wo.ab.CreateMetadata(w, r)
			log.Println(data)
			for wo.ab.ProposalState() != COMMITTED {
				// Propose in sequence to ensure Linearization Write
				time.Sleep(time.Second)
			}
			wo.ab.startProposal(data)
			return
		}
	})
}
