package database

import (
	"app/internal/database/models"
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const limit = 100

type driver struct {
	db     *gorm.DB
	logger *logrus.Logger
}

type Database interface {
	SaveTransaction(tx *models.Transaction) error
	GetTransactions(ctx context.Context, sender, receiver string, offset int) ([]models.Transaction, error)
}

func NewDriver(logger *logrus.Logger, dsn string) (Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	driverDB, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "postgres", driverDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return &driver{logger: logger, db: db}, nil
}

func (d *driver) SaveTransaction(tx *models.Transaction) error {
	return d.db.Transaction(func(dbTx *gorm.DB) error {
		result := dbTx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hash"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"status": tx.Status,
			}),
			Where: clause.Where{Exprs: []clause.Expression{
				clause.Eq{Column: "status", Value: "pending"},
			}},
		}).Create(tx)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("transaction exists and is not pending")
		}
		return nil
	})
}

func (d *driver) GetTransactions(ctx context.Context, sender, receiver string, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction

	db := d.db.WithContext(ctx)
	switch {
	case sender != "" && receiver != "":
		db = db.Where("sender = ? AND receiver = ?", sender, receiver)
	case sender != "":
		db = db.Where("sender = ?", sender)
	case receiver != "":
		db = db.Where("receiver = ?", receiver)
	}
	if err := db.Find(&transactions).Limit(limit).Offset(offset).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
