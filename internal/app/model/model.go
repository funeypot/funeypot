package model

import (
	"time"

	"github.com/gochore/boltutil"
)

type Record struct {
	Ip         string    `json:"ip"`
	StartedAt  time.Time `json:"started_at"`
	StoppedAt  time.Time `json:"stopped_at"`
	Count      int       `json:"count"`
	ReportedAt time.Time `json:"reported_at"`
	Score      int       `json:"score"`
}

func (r *Record) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

var _ boltutil.Storable = (*Record)(nil)

func (r *Record) BoltBucket() []byte {
	return []byte("records")
}

func (r *Record) BoltKey() []byte {
	return []byte(r.Ip)
}
