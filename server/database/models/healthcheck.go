package models

type HealthCheck struct {
	Message    string `json:"message"`
	PortNumber string `json:"portNumber"`
}
