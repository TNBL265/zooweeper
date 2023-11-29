package ztree

import (
	"database/sql"
)

type ZNodeHandler interface {
	// Utils
	Connection() *sql.DB
	InitializeDB()

	// Getter
	AllMetadata() ([]*Metadata, error)
	ZNodeIdExists(nodeId int) (bool, error)
	GetHighestZNodeId() (int, error)
	GetLocalMetadata() (*Metadata, error)
	GetMetadatasGreaterThanZNodeId(highestZNodeId int) (Metadatas, error)
	GetClients(client string) ([]string, error)

	// Setter
	InsertMetadata(metadata Metadata) error
	UpsertMetadata(metadata Metadata) error
	InsertMetadataWithParentId(metadata Metadata, parentId int) error
	UpdateFirstLeader(Leader string) error
}
