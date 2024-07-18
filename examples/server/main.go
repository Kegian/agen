package main

import (
	"context"

	"github.com/Kegian/agen/examples/server/internal/api"
	"github.com/Kegian/agen/examples/server/internal/config"
	"github.com/Kegian/agen/examples/server/internal/generated/oapi"
	"github.com/Kegian/agen/examples/server/internal/generated/server"
	"github.com/Kegian/agen/examples/server/internal/repo"
	"github.com/Kegian/agen/examples/server/internal/repo/ch"
	"github.com/Kegian/agen/examples/server/internal/repo/pg"

	"github.com/Kegian/agen"
)

func main() {
	ctx := context.Background()

	err := agen.Init(agen.WithConfig(&config.Cfg))
	if err != nil {
		panic(err)
	}
	defer agen.Sync() //nolint

	if err := Run(ctx); err != nil {
		panic(err)
	}
}

func Run(ctx context.Context) error {
	pgdb, err := agen.InitPostgres(ctx)
	if err != nil {
		return err
	}
	defer pgdb.Close()

	chdb, err := agen.InitClickhouse(ctx)
	if err != nil {
		return err
	}
	defer chdb.Close()

	repo := repo.New(pg.New(pgdb), ch.New(chdb))

	service := api.NewService(repo)
	srv, err := oapi.NewServer(
		service,
		&api.SecurityHandler{},
		oapi.WithMiddleware(
			agen.LogMiddleware(api.UserIDKey),
			agen.RecoverMiddleware,
		),
		oapi.WithErrorHandler(agen.ErrorHandler),
	)
	if err != nil {
		return err
	}

	return server.New(config.Cfg.ServerAddr, srv).Run()
}
