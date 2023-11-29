package zooweeper

import (
	rp "github.com/tnbl265/zooweeper/request_processors"
)

type Server struct {
	Rp rp.RequestProcessor
}

func NewServer(dbPath string) *Server {
	return &Server{
		Rp: *rp.NewRequestProcessor(dbPath),
	}
}
