package models

type HealthCheck struct {
	Message    string `json:"message"`
	PortNumber string `json:"portNumber"`
}

type HealthCheckError struct {
	Error     error
	ErrorPort string
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
