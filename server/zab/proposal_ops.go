package zooweeper

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type ProposalOps struct {
	ab *AtomicBroadcast
}

func (po *ProposalOps) ProposeWrite(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")
	if clientPort != zNode.Leader {
		log.Fatalf("I only supposed to receive Propose Write from leader not from %s\n", clientPort)
	}

	data := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(data)

	log.Printf("%s Receive Propose Write from %s\n", zNode.NodeIp, clientPort)
	url := po.ab.BaseURL + ":" + zNode.Leader + "/acknowledgeProposal"
	_ = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
}

func (po *ProposalOps) AcknowledgeProposal(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")
	if clientPort == zNode.Leader {
		log.Fatalf("I'm the Leader I'm not supposed to get acknowledged from myself\n")
	}
	log.Printf("Received Acknowledgement from %s\n", clientPort)

	// Wait for all Follower to ACK
	portsSlice := strings.Split(zNode.Servers, ",")
	if po.ab.AckCounter() != len(portsSlice)-1 {
		counter := po.ab.AckCounter()
		counter++
		po.ab.SetAckCounter(counter)

		if po.ab.AckCounter() == len(portsSlice)-1 {
			po.ab.SetAckCounter(0)
			po.ab.SetProposalState(ACKNOWLEDGED)
		}
	}

	data := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(data)

	log.Printf("Asking Follower %s to commit\n", clientPort)
	url := po.ab.BaseURL + ":" + clientPort + "/commitWrite"
	_ = po.ab.makeExternalRequest(nil, url, "POST", jsonData)

}

func (po *ProposalOps) CommitWrite(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()

	data := po.ab.CreateMetadata(w, r)

	jsonData, _ := json.Marshal(data)
	log.Printf("%s Commiting Write\n", zNode.NodeIp)
	url := po.ab.BaseURL + ":" + zNode.NodeIp + "/writeMetadata"
	_ = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
}
