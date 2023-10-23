package zooweeper

import (
	"database/sql"

	"github.com/tnbl265/zooweeper/server/database/models"
)

type ZooWeeperDatabaseRepo interface {
	Connection() *sql.DB
	InsertMetadata(metadata models.Metadata) error
	AllMetadata() ([]*models.Metadata, error)
}
