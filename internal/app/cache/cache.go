package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sshless/internal/app/model"
	"sshless/internal/pkg/logs"

	"github.com/dgraph-io/badger/v4"
)

type Cache struct {
	db *badger.DB
}

func New(ctx context.Context, dir string) (*Cache, error) {
	options := badger.DefaultOptions(dir)
	options.Logger = logs.NewBadgerLogger(logs.From(ctx))
	db, err := badger.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open badger: %w", err)
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Close() {
	_ = c.db.Close()
}

func (c *Cache) GetRecord(_ context.Context, ip string) (*model.Record, bool, error) {
	record := &model.Record{}
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(ip))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, record)
		})
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get record: %w", err)
	}
	return record, true, nil
}

func (c *Cache) SetRecord(_ context.Context, record *model.Record) error {
	return c.db.Update(func(txn *badger.Txn) error {
		val, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry([]byte(record.Ip), val).WithTTL(24 * time.Hour))
	})
}

func (c *Cache) IncrRecord(ctx context.Context, ip string) (*model.Record, error) {
	record, ok, err := c.GetRecord(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("get record: %v", err)
	}
	if !ok {
		record = &model.Record{
			Ip:        ip,
			StartedAt: time.Now(),
		}
	}
	record.Count++
	record.StartedAt = time.Now()
	if err := c.SetRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("set record: %v", err)
	}
	return record, nil
}

func (c *Cache) AllRecords(ctx context.Context) ([]*model.Record, error) {
	records := make([]*model.Record, 0)
	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			if err := ctx.Err(); err != nil {
				return err
			}
			item := it.Item()
			record := &model.Record{}
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, record)
			}); err != nil {
				return err
			}
			records = append(records, record)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get all records: %w", err)
	}
	return records, nil
}
