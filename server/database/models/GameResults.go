package models

import "time"

type GameResults struct {
	Minute int    `json:"Minute"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

type UpdateScore struct {
	SenderIp   string       `json:"SenderIp"`
	ReceiverIp string       `json:"ReceiverIp"`
	Timestamp  time.Time    `json:"Timestamp"`
	Attempts   int          `json:"Attempts"`
	Event      *GameResults `json:"Event"`
}
