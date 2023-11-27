package zooweeper

import (
	"database/sql"

	"github.com/tnbl265/zooweeper/database/models"
)

type ZooWeeperDatabaseRepo interface {
	// Utils
	Connection() *sql.DB
	InitializeDB()
	NodeIdExists(nodeId int) (bool, error)
	GetLocalMetadata() (*models.Metadata, error)

	// ZNode
	GetClients(client string) ([]string, error)
	InsertMetadata(metadata models.Metadata, parentId int) error
	UpsertMetadata(metadata models.Metadata) error
	AllMetadata() ([]*models.Metadata, error)

	UpdateFirstLeader(Leader string) error
}
