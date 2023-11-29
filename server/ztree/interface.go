package zooweeper

import (
	"database/sql"

	"github.com/tnbl265/zooweeper/database/models"
)

type ZNodeHandler interface {
	// Utils
	Connection() *sql.DB
	InitializeDB()

	// Getter
	AllMetadata() ([]*models.Metadata, error)
	ZNodeIdExists(nodeId int) (bool, error)
	GetHighestZNodeId() (int, error)
	GetLocalMetadata() (*models.Metadata, error)
	GetMetadatasGreaterThanZNodeId(highestZNodeId int) (models.Metadatas, error)
	GetClients(client string) ([]string, error)

	// Setter
	InsertMetadata(metadata models.Metadata) error
	UpsertMetadata(metadata models.Metadata) error
	InsertMetadataWithParentId(metadata models.Metadata, parentId int) error
	UpdateFirstLeader(Leader string) error
}
