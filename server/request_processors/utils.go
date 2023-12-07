package request_processors

import "net/http"

// Transaction with Timestamp field as id for ordering in PriorityQueue
type Transaction struct {
	Request   *http.Request
	Timestamp string
}

// PriorityQueue of Transaction to be used by QueueMiddleware
type PriorityQueue []*Transaction

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Timestamp < pq[j].Timestamp
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Transaction)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Peek() *Transaction {
	if len(*pq) == 0 {
		return nil
	}
	return (*pq)[0]
}
