package zooweeper

import (
	"database/sql"
	"fmt"
	ztree "github.com/tnbl265/zooweeper/database"
	"github.com/tnbl265/zooweeper/database/handlers"
	"log"
	"net/http"
)

// Self-reference to parent - ref: https://stackoverflow.com/questions/27918208/go-get-parent-struct
type AtomicBroadcast struct {
	Read  ReadOps
	Write WriteOps
	ZTree ztree.ZooWeeperDatabaseRepo
}

func NewAtomicBroadcast(dbPath string) *AtomicBroadcast {
	ab := &AtomicBroadcast{}

	// Connect to the Database
	log.Println("Connecting to", dbPath)
	db, err := ab.OpenDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	ab.ZTree = &handlers.ZTree{DB: db}
	ab.Read.ab = ab
	ab.Write.ab = ab

	ab.ZTree.InitializeDB()
	return ab
}

func (ab *AtomicBroadcast) Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong")
}

func (ab *AtomicBroadcast) OpenDB(datasource string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", datasource)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
