package zooweeper

import (
	"encoding/json"
	"log"
	"net/http"
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

	metadata := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(metadata)

	//log.Printf("%s Receive Propose Write from %s\n", zNode.NodeIp, clientPort)
	url := "http://localhost:" + zNode.Leader + "/acknowledgeProposal"
	_ = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
}

func (po *ProposalOps) AcknowledgeProposal(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")
	if clientPort == zNode.Leader {
		log.Fatalf("I'm the Leader I'm not supposed to get acknowledged from myself\n")
	}

	metadata := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(metadata)

	//log.Printf("Received Acknowledgement from %s\n", clientPort)
	//log.Printf("Asking Follower %s to commit\n", clientPort)
	url := "http://localhost:" + clientPort + "/commitWrite"
	_ = po.ab.makeExternalRequest(nil, url, "POST", jsonData)

}

func (po *ProposalOps) CommitWrite(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()

	metadata := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(metadata)

	//log.Printf("%s Commiting Write\n", zNode.NodeIp)
	url := "http://localhost:" + zNode.NodeIp + "/writeMetadata"
	_ = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
}
