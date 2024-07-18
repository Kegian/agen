package agen

import (
	"context"

	"github.com/Kegian/agen/database"
)

func InitPostgres(ctx context.Context) (*database.PostgresDB, error) {
	pg := database.NewPostgres(&Config.Postgres)
	if err := pg.Open(ctx); err != nil {
		return nil, err
	}
	return pg, nil
}

func InitClickhouse(ctx context.Context) (*database.ClickhouseDB, error) {
	ch := database.NewClickhouse(&Config.Clickhouse)
	if err := ch.Open(ctx); err != nil {
		return nil, err
	}
	return ch, nil
}
