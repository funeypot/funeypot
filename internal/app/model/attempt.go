package model

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

func init() {
	registerModel(new(BruteAttempt))
}

type BruteAttemptKind int

const (
	_ BruteAttemptKind = iota
	BruteAttemptKindSsh
	BruteAttemptKindHttp
)

type BruteAttempt struct {
	Id            int64
	Ip            string           `gorm:"size:39,index:ip_kind"` // max ipv6 length
	Kind          BruteAttemptKind `gorm:"index:ip_kind"`
	User          string           `gorm:"size:255"`
	Password      string           `gorm:"size:255"`
	ClientVersion string           `gorm:"size:255"`
	StartedAt     time.Time
	StoppedAt     time.Time
	Count         int64

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"index"`
}

func (r *BruteAttempt) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

func (r *BruteAttempt) MaskedPassword() string {
	prefix := len(r.Password) / 3
	if prefix > 4 {
		prefix = 4
	}
	suffix := prefix

	return r.Password[:prefix] + strings.Repeat("*", len(r.Password)-prefix-suffix) + r.Password[len(r.Password)-suffix:]
}

func (r *BruteAttempt) ShortClientVersion() string {
	return strings.TrimPrefix(r.ClientVersion, "SSH-2.0-")
}

func (r *BruteAttempt) BeforeSave(_ *gorm.DB) error {
	r.User = truncateString(r.User, 255)
	r.Password = truncateString(r.Password, 255)
	r.ClientVersion = truncateString(r.ClientVersion, 255)
	return nil
}

func (db *Database) LastBruteAttempt(ctx context.Context, ip string, kind BruteAttemptKind) (*BruteAttempt, bool, error) {
	attempt := &BruteAttempt{}
	result := db.db.
		WithContext(ctx).
		Last(&attempt, "ip = ? AND kind = ?", ip, kind)
	if err := result.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return attempt, true, nil
}

func (db *Database) ScanBruteAttempt(ctx context.Context, updatedAfter time.Time, f func(attempt *BruteAttempt, geo *IpGeo) bool) error {
	rows, err := db.db.
		WithContext(ctx).
		Model(&BruteAttempt{}).
		Select("brute_attempt.*, ip_geos.*").
		Joins("LEFT JOIN ip_geos ON brute_attempt.ip = ip_geos.ip").
		Where("brute_attempt.updated_at > ?", updatedAfter).
		Order("brute_attempt.updated_at").
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		result := &struct {
			*BruteAttempt
			*IpGeo
		}{}
		if err := db.db.ScanRows(rows, result); err != nil {
			return err
		}
		if !f(result.BruteAttempt, result.IpGeo) {
			break
		}
	}
	return nil
}
