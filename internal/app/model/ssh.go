package model

import (
	"strings"
	"time"
)

func init() {
	registerModel(new(SshAttempt))
}

type SshAttempt struct {
	Id            int64  `gorm:"primaryKey autoIncrement"`
	Ip            string `gorm:"size:39"` // max ipv6 length
	User          string `gorm:"size:255"`
	Password      string `gorm:"size:255"`
	ClientVersion string `gorm:"size:32"`
	StartedAt     time.Time
	StoppedAt     time.Time
	Count         int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *SshAttempt) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

func (r *SshAttempt) MaskedPassword() string {
	prefix := len(r.Password) / 3
	if prefix > 4 {
		prefix = 4
	}
	suffix := prefix

	return r.Password[:prefix] + strings.Repeat("*", len(r.Password)-prefix-suffix) + r.Password[len(r.Password)-suffix:]
}

func (r *SshAttempt) ShortClientVersion() string {
	return strings.TrimPrefix(r.ClientVersion, "SSH-2.0-")
}
