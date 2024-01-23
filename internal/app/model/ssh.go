package model

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

func init() {
	registerModel(new(SshAttempt))
}

type SshAttempt struct {
	Id            int64
	Ip            string `gorm:"size:39"` // max ipv6 length
	User          string `gorm:"size:255"`
	Password      string `gorm:"size:255"`
	ClientVersion string `gorm:"size:32"`
	StartedAt     time.Time
	StoppedAt     time.Time
	Count         int64

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"index"`
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

func (db *Database) LastSshAttempt(ctx context.Context, ip string) (*SshAttempt, bool, error) {
	attempt := &SshAttempt{}
	result := db.db.
		WithContext(ctx).
		Last(&attempt, "ip = ?", ip)
	if err := result.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return attempt, true, nil
}

func (db *Database) ScanSshAttempt(ctx context.Context, updatedAfter time.Time, f func(attempt *SshAttempt, geo *IpGeo) bool) error {
	rows, err := db.db.
		WithContext(ctx).
		Model(&SshAttempt{}).
		Select("ssh_attempts.*, ip_geos.*").
		Joins("LEFT JOIN ip_geos ON ssh_attempts.ip = ip_geos.ip").
		Where("ssh_attempts.updated_at > ?", updatedAfter).
		Order("ssh_attempts.updated_at").
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		result := &struct {
			*SshAttempt
			*IpGeo
		}{}
		if err := db.db.ScanRows(rows, result); err != nil {
			return err
		}
		if !f(result.SshAttempt, result.IpGeo) {
			break
		}
	}
	return nil
}
