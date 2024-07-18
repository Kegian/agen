package repo

import (
	"errors"

	"github.com/Kegian/agen/examples/server/internal/repo/ch"
	"github.com/Kegian/agen/examples/server/internal/repo/pg"

	"github.com/jackc/pgx/v5"
)

type Repo struct {
	*pg.PostgresDB
	*ch.ClickhouseDB
}

func New(pg *pg.PostgresDB, ch *ch.ClickhouseDB) *Repo {
	return &Repo{
		PostgresDB:   pg,
		ClickhouseDB: ch,
	}
}

func (r *Repo) IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
