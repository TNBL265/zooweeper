package zooweeper

import "github.com/tnbl265/zooweeper/request_processors"

type Server struct {
	Rp request_processors.RequestProcessor
}

func NewServer(dbPath string) *Server {
	return &Server{
		Rp: *request_processors.NewRequestProcessor(dbPath),
	}
}
