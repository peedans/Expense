package expenses

import (
	"database/sql"
	"fmt"
)

type Handler struct {
	DB *sql.DB
}

func InitDB(db *sql.DB) *Handler {

	createTb := `CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);`

	_, err := db.Exec(createTb)
	if err != nil {
		fmt.Println("can't create table", err)
	}

	return &Handler{db}
}
