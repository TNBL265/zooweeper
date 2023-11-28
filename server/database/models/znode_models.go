package models

type Metadata struct {
	NodeId     int    `json:"NodeId"`
	NodeIp     string `json:"NodeIp"`
	Leader     string `json:"Leader"`
	Servers    string `json:"Servers"`
	Timestamp  string `json:"Timestamp"`
	Attempts   int    `json:"Attempts"`
	Version    int    `json:"Version"`
	ParentId   int    `json:"ParentId"`
	Clients    string `json:"Clients"`
	SenderIp   string `json:"SenderIp"`
	ReceiverIp string `json:"ReceiverIp"`
}

type GameResults struct {
	Minute int    `json:"Minute"`
	Player string `json:"Player"`
	Club   string `json:"Club"`
	Score  string `json:"Score"`
}

type Data struct {
	Timestamp   string      `json:"Timestamp"`
	Metadata    Metadata    `json:"Metadata"`
	GameResults GameResults `json:"GameResults"`
}

type Metadatas struct {
	MetadataList []Metadata `json:"MetadataList"`
}
