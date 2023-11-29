// Package data defines json schema for data of POST request
//
// 1. External POST request: from client (Kafka-Server) to query/update Metadata
// 2. Internal POST request: from ZooWeeper servers for HealthCheck and LeaderElection

package data

import (
	"github.com/tnbl265/zooweeper/ztree"
)

type GameResults struct {
	Minute int    `json:"Minute"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

type Data struct {
	Timestamp   string         `json:"Timestamp"`
	Metadata    ztree.Metadata `json:"Metadata"`
	GameResults GameResults    `json:"GameResults"`
}

type HealthCheck struct {
	Message    string `json:"message"`
	PortNumber string `json:"portNumber"`
}

type HealthCheckError struct {
	Error     error
	ErrorPort string
	IsWakeup  bool
}

type ElectLeaderRequest struct {
	IncomingPort string `json:"port"`
}

type ElectLeaderResponse struct {
	IsSuccess string `json:"isSuccess"`
}
type DeclareLeaderRequest struct {
	IncomingPort string `json:"port"`
}
