package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Record struct {
	Ip         string
	StartedAt  time.Time
	StoppedAt  time.Time
	Count      int
	Geo        string
	ReportedAt time.Time
}

func (r *Record) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

var (
	records = make(map[string]*Record)
	recordM sync.Mutex
)

func GetRecord(ctx context.Context, ip string) Record {
	recordM.Lock()
	defer recordM.Unlock()

	for k, v := range records {
		if time.Since(v.StoppedAt) > 24*time.Hour {
			delete(records, k)
			continue
		}
	}

	now := time.Now()

	record, ok := records[ip]
	if !ok {
		record = &Record{
			Ip:        ip,
			StartedAt: now,
		}
		records[ip] = record
	}

	record.StoppedAt = now
	record.Count++

	if record.Geo == "" {
		geo, err := IpGeo(ctx, ip)
		if err != nil {
			slog.Error("get ip geo",
				"ip", ip,
				"error", err,
			)
		} else {
			record.Geo = geo
		}
	}

	return *record
}
