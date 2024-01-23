package model

import (
	"time"
)

func init() {
	registerModel(new(AbuseipdbReport))
}

type AbuseipdbReport struct {
	Id         int64
	Ip         string `gorm:"size:39"`
	ReportedAt time.Time
	Score      int

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}
