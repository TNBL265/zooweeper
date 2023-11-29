package zooweeper

import "github.com/tnbl265/zooweeper/ztree/models"

type GameResults struct {
	Minute int    `json:"Minute"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

type Data struct {
	Timestamp   string          `json:"Timestamp"`
	Metadata    models.Metadata `json:"Metadata"`
	GameResults GameResults     `json:"GameResults"`
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
