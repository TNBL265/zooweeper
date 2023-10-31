package zooweeper

import (
	rp "github.com/tnbl265/zooweeper/request_processors"
	"log"
)

type ServerState string

const (
	LEADING   ServerState = "LEADING"
	FOLLOWING ServerState = "FOLLOWING"
	//ELECTING
)

type Server struct {
	nodeId     int
	leader     int
	allServers []int
	state      ServerState
	Rp         rp.RequestProcessor
}

func (svr *Server) NodeId() int {
	return svr.nodeId
}

func (svr *Server) SetNodeId(nodeId int) {
	svr.nodeId = nodeId
}

func (svr *Server) Leader() int {
	return svr.leader
}

func (svr *Server) SetLeader(leader int) {
	svr.leader = leader
}

func (svr *Server) AllServers() []int {
	return svr.allServers
}

func (svr *Server) SetAllServers(allServers []int) {
	svr.allServers = allServers
}

func (svr *Server) State() ServerState {
	return svr.state
}

func (svr *Server) SetState(state ServerState) {
	svr.state = state
}

func NewServer(port, leader int, allServers []int) *Server {
	server := &Server{
		nodeId:     port,
		leader:     leader,
		allServers: allServers,
	}

	if port == 8080 {
		server.state = LEADING
	} else if port == 8081 || port == 8082 {
		server.state = FOLLOWING
	} else {
		log.Fatalf("Only support ports 8080, 8081 or 8082")
	}
	return server
}
