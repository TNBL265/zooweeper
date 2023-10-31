package zooweeper

import (
	rp "github.com/tnbl265/zooweeper/request_processors"
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

func NewServer(port, leader int, state ServerState, allServers []int, dbPath string) *Server {
	return &Server{
		nodeId:     port,
		leader:     leader,
		state:      state,
		allServers: allServers,
		Rp:         *rp.NewRequestProcessor(dbPath),
	}
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
