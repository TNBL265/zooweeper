package zooweeper

import "database/sql"

func (app *Application) OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "server/database/zooweeper-database.db")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
