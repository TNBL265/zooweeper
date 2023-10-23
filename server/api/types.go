package zooweeper

import zooweeper "github.com/tnbl265/zooweeper/server/database"

type Application struct {
	Domain string
	DB     zooweeper.ZooWeeperDatabaseRepo
}
