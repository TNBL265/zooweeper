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
	GetServers() ([]string, error)
	InsertMetadata(metadata models.Metadata) error
	AllMetadata() ([]*models.Metadata, error)
	DeleteMetadata(Leader string) error
	CheckMetadataExist(Leader string) (bool, error)
}
