package zooweeper

import (
	"database/sql"

	"github.com/tnbl265/zooweeper/database/models"
)

type ZooWeeperDatabaseRepo interface {
	// Utils
	Connection() *sql.DB
	InitializeDB()
	ZNodeIdExists(nodeId int) (bool, error)
	GetLocalMetadata() (*models.Metadata, error)
	GetHighestZNodeId() (int, error)
	GetMetadatasGreaterThanZNodeId(highestZNodeId int) (models.Metadatas, error)

	// ZNode
	GetClients(client string) ([]string, error)
	InsertMetadata(metadata models.Metadata) error
	UpsertMetadata(metadata models.Metadata) error
	InsertMetadataWithParentId(metadata models.Metadata, parentId int) error
	AllMetadata() ([]*models.Metadata, error)
	UpdateFirstLeader(Leader string) error
}
