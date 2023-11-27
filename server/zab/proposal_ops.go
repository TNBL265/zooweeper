package zooweeper

import (
	"encoding/json"
	"github.com/fatih/color"
	"net/http"
	"strings"
	"time"
)

type ProposalOps struct {
	ab *AtomicBroadcast
}

func (po *ProposalOps) ProposeWrite(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")
	if clientPort != zNode.Leader {
		color.Red("I only supposed to receive Propose Write from leader not from %s\n", clientPort)
	}

	data := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(data)

	color.HiBlue("%s received Propose Write from %s\n", zNode.NodeIp, clientPort)
	color.HiBlue("%s sending ACK to %s\n", zNode.NodeIp, clientPort)
	url := po.ab.BaseURL + ":" + zNode.Leader + "/acknowledgeProposal"
	_, err = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
}

func (po *ProposalOps) AcknowledgeProposal(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")
	if clientPort == zNode.Leader {
		color.Red("I'm the Leader I'm not supposed to get acknowledged from myself\n")
	}
	color.HiBlue("Leader %s received ACK from Follower %s\n", zNode.NodeIp, clientPort)

	if po.ab.ProposalState() != ACKNOWLEDGED {
		currentAckCount := po.ab.AckCounter()
		currentAckCount++
		po.ab.SetAckCounter(currentAckCount)
		// Wait for majority of Follower to ACK
		portsSlice := strings.Split(zNode.Servers, ",")
		majority := len(portsSlice) / 2
		for {
			if currentAckCount > majority {
				color.HiBlue("Leader %s received majority ACK, %d\n", zNode.NodeIp, currentAckCount)
				po.ab.SetAckCounter(0)
				po.ab.SetProposalState(ACKNOWLEDGED)
				break
			} else if po.ab.ProposalState() == ACKNOWLEDGED {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	data := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(data)

	color.HiBlue("Leader %s asking Follower %s to commit\n", zNode.NodeIp, clientPort)
	url := po.ab.BaseURL + ":" + clientPort + "/commitWrite"
	_, err = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
	if err != nil {
		color.HiBlue("Error Asking Follower %s to commit: %s\n", clientPort, err.Error())
	}

}

func (po *ProposalOps) CommitWrite(w http.ResponseWriter, r *http.Request) {
	zNode, _ := po.ab.ZTree.GetLocalMetadata()
	clientPort := r.Header.Get("X-Sender-Port")

	data := po.ab.CreateMetadata(w, r)
	jsonData, _ := json.Marshal(data)

	color.HiBlue("%s receive Commit Write from %s\n", zNode.NodeIp, clientPort)
	color.HiBlue("%s Committing Write\n", zNode.NodeIp)
	url := po.ab.BaseURL + ":" + zNode.NodeIp + "/writeMetadata"
	_, err = po.ab.makeExternalRequest(nil, url, "POST", jsonData)
	if err != nil {
		color.Red("Error Commiting Write: %s\n", err.Error())
	}
}
