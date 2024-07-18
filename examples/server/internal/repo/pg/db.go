package pg

import (
	"github.com/Kegian/agen/database"
	"github.com/Kegian/agen/examples/server/internal/generated/query"
)

type PostgresDB struct {
	*query.Queries
	DB *database.PostgresDB
}

func New(db *database.PostgresDB) *PostgresDB {
	return &PostgresDB{
		Queries: query.New(db),
		DB:      db,
	}
}
