package model

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/wolfogre/funeypot/internal/app/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg config.Database) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	switch cfg.Driver {
	case sqlite.DriverName:
		db, err = gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // TODO: use logs
		})
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(models...); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

var models []any

func registerModel(model any) {
	for _, m := range models {
		if fmt.Sprintf("%T", m) == fmt.Sprintf("%T", model) {
			panic(fmt.Sprintf("model %T already registered", model))
		}
	}
	models = append(models, model)
}
