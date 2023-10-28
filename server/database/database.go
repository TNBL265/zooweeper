package zooweeper

import (
	"database/sql"

	"github.com/tnbl265/zooweeper/database/models"
)

type ZooWeeperDatabaseRepo interface {
	Connection() *sql.DB
	InsertMetadata(metadata models.Sello) error
	AllMetadata() ([]*models.Sello, error)
	DeleteMetadata(LeaderServer string) error
	CheckMetadataExist(LeaderServer string) (bool, error)
}
