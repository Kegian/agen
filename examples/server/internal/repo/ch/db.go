package ch

import (
	"context"
	"time"

	"github.com/Kegian/agen/database"

	"github.com/google/uuid"
)

type ClickhouseDB struct {
	*database.ClickhouseDB
}

func New(db *database.ClickhouseDB) *ClickhouseDB {
	return &ClickhouseDB{ClickhouseDB: db}
}

func (db *ClickhouseDB) SendStat(ctx context.Context, userID string, eventName string, valI int32, valS string) error {
	batch, err := db.PrepareBatch(ctx, "INSERT INTO analytics_log")
	if err != nil {
		return err
	}

	err = db.AsyncInsert(
		ctx,
		"INSERT INTO analytics_log VALUES (?, ?, ?, ?, ?, ?)",
		false,
		uuid.New().String(),
		userID,
		eventName,
		valI,
		valS,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return batch.Send()
}
