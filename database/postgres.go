package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresConfig struct {
	Host     string `cfg:"PG_HOST" default:"localhost"`
	Port     uint16 `cfg:"PG_PORT" default:"5432"`
	User     string `cfg:"PG_USER" default:"postgres"`
	Pass     string `cfg:"PG_PASS" default:"postgres"`
	Name     string `cfg:"PG_NAME" default:"postgres"`
	MaxConns uint16 `cfg:"PG_MAX_CONNS" default:"10"`
}

func NewPostgres(cfg *PostgresConfig) *PostgresDB {
	db := &PostgresDB{
		Addr:     cfg.Host + ":" + strconv.Itoa(int(cfg.Port)),
		User:     cfg.User,
		Pass:     cfg.Pass,
		Name:     cfg.Name,
		MaxConns: cfg.MaxConns,
	}
	db.Logger = zap.L().With(
		zap.Namespace("postgres"),
		zap.String("addr", db.Addr),
		zap.String("name", db.Name),
	)

	return db
}

type PostgresDB struct {
	*pgxpool.Pool

	Addr     string
	User     string
	Pass     string
	Name     string
	MaxConns uint16

	Logger *zap.Logger
}

func (db *PostgresDB) Open(ctx context.Context) (err error) {
	db.Logger.Info("connecting to postgres database")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?pool_max_conns=%d",
		db.User,
		db.Pass,
		db.Addr,
		db.Name,
		db.MaxConns,
	)

	db.Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		db.Logger.Error("postgres database connection failed", zap.Error(err))
		return err
	}

	err = db.Ping(ctx)
	if err != nil {
		db.Logger.Error("postgres database connection failed", zap.Error(err))
		return err
	}

	db.Logger.Info("successfully connected to postgres database")

	return nil
}

func (db *PostgresDB) Close() {
	db.Pool.Close()
}
