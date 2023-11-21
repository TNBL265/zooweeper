package models

type HealthCheck struct {
	Message    string `json:"message"`
	PortNumber string `json:"portNumber"`
}

type HealthCheckError struct {
	Error     error
	ErrorPort string
}
