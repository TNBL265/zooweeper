package zooweeper

import zooweeper "github.com/tnbl265/zooweeper/database"

type Application struct {
	Domain string
	DB     zooweeper.ZooWeeperDatabaseRepo
}
