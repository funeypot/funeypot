package model

import (
	"context"
	"time"

	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/gochore/boltutil"
)

func NewDatabase(ctx context.Context, file string) (*boltutil.DB, error) {
	db, err := boltutil.Open(
		file,
		boltutil.WithDefaultCoder(boltutil.JsonCoder{}),
	)
	if err != nil {
		return nil, err
	}
	go gc(ctx, db)
	return db, nil
}

func gc(ctx context.Context, db *boltutil.DB) {
	logger := logs.From(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Info("gc done")
			return
		case <-time.After(time.Minute):
		}

		var records []*Record
		if err := db.Scan(&records, boltutil.NewFilter().AddStorableCondition(func(obj boltutil.Storable) (skip bool, stop bool) {
			record := obj.(*Record)
			if time.Since(record.StoppedAt) > time.Hour*24 {
				return false, false
			}
			return true, false
		})); err != nil {
			logger.Errorf("scan: %v", err)
			continue
		}

		for _, record := range records {
			logger.Debugf("delete %v", record)
			if err := db.Delete(record); err != nil {
				logger.Errorf("delete: %v", err)
			}
		}
	}
}
