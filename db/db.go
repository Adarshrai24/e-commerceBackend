package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var DB *sql.DB
