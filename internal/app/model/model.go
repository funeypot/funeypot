package model

import (
	"time"
)

type Record struct {
	Ip         string    `json:"ip"`
	StartedAt  time.Time `json:"started_at"`
	StoppedAt  time.Time `json:"stopped_at"`
	Count      int       `json:"count"`
	Geo        string    `json:"geo"`
	ReportedAt time.Time `json:"reported_at"`
}

func (r *Record) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}
