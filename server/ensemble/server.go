// Package ensemble implements a ZooWeeper ensemble.
//
// 1. It defines the Server struct that represents a ZooWeeper server.
// 2. The Server struct has access to its local ZTree database via RequestProcessor.Zab.ZTree

package ensemble

import "github.com/tnbl265/zooweeper/request_processors"

type Server struct {
	Rp request_processors.RequestProcessor
}

func NewServer(dbPath string) *Server {
	return &Server{
		Rp: *request_processors.NewRequestProcessor(dbPath),
	}
}
