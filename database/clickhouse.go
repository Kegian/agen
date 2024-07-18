package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Kegian/agen/errors"
	"github.com/Kegian/agen/sentry"
	"go.uber.org/zap"
)

type ClickhouseConfig struct {
	Host        string `cfg:"CH_HOST" default:"localhost"`
	Port        uint16 `cfg:"CH_PORT" default:"9000"`
	User        string `cfg:"CH_USER" default:"default"`
	Pass        string `cfg:"CH_PASS" default:"default"`
	Name        string `cfg:"CH_NAME" default:"default"`
	MustConnect bool   `cfg:"CH_MUST_CONNECT" default:"false"`
}

func NewClickhouse(cfg *ClickhouseConfig) *ClickhouseDB {
	db := &ClickhouseDB{
		Addr: cfg.Host + ":" + strconv.Itoa(int(cfg.Port)),
		User: cfg.User,
		Pass: cfg.Pass,
		Name: cfg.Name,
		Must: cfg.MustConnect,
	}
	db.Logger = zap.L().With(
		zap.Namespace("clickhouse"),
		zap.String("addr", db.Addr),
		zap.String("name", db.Name),
	)

	return db
}

type ClickhouseDB struct {
	driver.Conn

	Addr string
	User string
	Pass string
	Name string
	Must bool

	Logger *zap.Logger
}

func (db *ClickhouseDB) Open(ctx context.Context) (err error) {
	if db.Must {
		return db.open(ctx)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := db.open(ctx); err != nil {
				db.Logger.Error("error on clickhouse connect", zap.Error(err))
				sentry.SentryCaptureException(ctx, errors.Wrap(err, "error on clickhouse connect"), nil)
				time.Sleep(5 * time.Second)
				continue
			}
			return
		}
	}()

	return nil
}

func (db *ClickhouseDB) Close() error {
	return db.Conn.Close()
}

func (db *ClickhouseDB) open(ctx context.Context) (err error) {
	db.Logger.Info("connecting to clickhouse database")

	db.Conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{db.Addr},
		Auth: clickhouse.Auth{
			Database: db.Name,
			Username: db.User,
			Password: db.Pass,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "agen", Version: "0.1"},
			},
		},

		Debugf: func(format string, v ...interface{}) {
			db.Logger.Debug(fmt.Sprintf(format, v))
		},
	})

	if err != nil {
		db.Logger.Error("clickhouse database connection failed", zap.Error(err))
		return err
	}

	if err := db.Ping(ctx); err != nil {
		db.Logger.Error("clickhouse database connection failed", zap.Error(err))
		return err
	}

	db.Logger.Info("successfully connected to clickhouse database")
	return nil
}
