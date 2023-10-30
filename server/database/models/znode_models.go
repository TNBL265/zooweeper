package models

import "time"

type Metadata struct {
	LeaderServer string    `json:"LeaderServer"`
	Servers      string    `json:"Servers"`
	SenderIp     string    `json:"SenderIp"`
	ReceiverIp   string    `json:"ReceiverIp"`
	Timestamp    time.Time `json:"Timestamp"`
	Attempts     int       `json:"Attempts"`
}

type GameResults struct {
	Minute int    `json:"Minute"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

type Data struct {
	Metadata
	GameResults
}
