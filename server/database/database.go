package zooweeper

import (
	"database/sql"

	"github.com/tnbl265/zooweeper/server/database/models"
)

type ZooWeeperDatabaseRepo interface {
	Connection() *sql.DB
	AllGameResults() ([]*models.GameResults, error)
}
