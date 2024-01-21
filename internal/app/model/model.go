package model

import (
	"strings"
	"time"

	"github.com/gochore/boltutil"
)

type Record struct {
	Ip            string    `json:"ip"`
	User          string    `json:"user"`
	Password      string    `json:"password"`
	ClientVersion string    `json:"client_version"`
	StartedAt     time.Time `json:"started_at"`
	StoppedAt     time.Time `json:"stopped_at"`
	Count         int       `json:"count"`
	ReportedAt    time.Time `json:"reported_at"`
	Score         int       `json:"score"`
}

func (r *Record) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

func (r *Record) MaskedPassword() string {
	prefix := len(r.Password) / 3
	if prefix > 4 {
		prefix = 4
	}
	suffix := prefix

	return r.Password[:prefix] + strings.Repeat("*", len(r.Password)-prefix-suffix) + r.Password[len(r.Password)-suffix:]
}

func (r *Record) ShortClientVersion() string {
	return strings.TrimPrefix(r.ClientVersion, "SSH-2.0-")
}

var _ boltutil.Storable = (*Record)(nil)

func (r *Record) BoltBucket() []byte {
	return []byte("records")
}

func (r *Record) BoltKey() []byte {
	return []byte(r.Ip)
}
