package models

import "time"

type GameResults struct {
	Minute int    `json:"Min"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

type Metadata struct {
	SenderIp   string    `json:"SenderIp"`
	ReceiverIp string    `json:"ReceiverIp"`
	Timestamp  time.Time `json:"Timestamp"`
	Attempts   int       `json:"Attempts"`
}

type ServersData struct {
	LeaderServer string `json:"LeaderServer"`
	Servers      string `json:"Servers"`
}

type Sello struct {
	ServersData
	Metadata
}

type Score struct {
	ServersData
	Metadata
	Event GameResults `json:"Event"`
}
