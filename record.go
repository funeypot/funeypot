package main

import (
	"sync"
	"time"
)

type Record struct {
	StartedAt time.Time
	StoppedAt time.Time
	Count     int
}

func (r *Record) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

var (
	records = make(map[string]*Record)
	recordM sync.Mutex
)

func GetRecord(ip string) Record {
	recordM.Lock()
	defer recordM.Unlock()

	for k, v := range records {
		if time.Since(v.StoppedAt) > 24*time.Hour {
			delete(records, k)
			continue
		}
	}

	now := time.Now()

	if record, ok := records[ip]; ok {
		record.StoppedAt = now
		record.Count++
		return *record
	}

	record := &Record{
		StartedAt: now,
		StoppedAt: now,
		Count:     1,
	}

	records[ip] = record
	return *record
}
