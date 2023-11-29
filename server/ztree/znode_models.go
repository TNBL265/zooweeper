package ztree

type Metadata struct {
	NodeId     int    `json:"NodeId"`
	NodeIp     string `json:"NodeIp"`
	Leader     string `json:"Leader"`
	Servers    string `json:"Servers"`
	Timestamp  string `json:"Timestamp"`
	Version    int    `json:"Version"`
	ParentId   int    `json:"ParentId"`
	Clients    string `json:"Clients"`
	SenderIp   string `json:"SenderIp"`
	ReceiverIp string `json:"ReceiverIp"`
}

type Metadatas struct {
	MetadataList []Metadata `json:"MetadataList"`
}
