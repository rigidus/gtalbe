package database

import (
	"app/internal/database/models"
	"context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const limit = 100

type driver struct {
	db *gorm.DB
}

type Database interface {
	SaveTransaction(tx *models.Transaction) error
	GetTransactions(ctx context.Context, sender, receiver string) ([]models.Transaction, error)
}

func NewDriver(dsn string) (Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// if err := db.AutoMigrate(&Transaction{}); err != nil {
	//		log.Fatal("migration failed:", err)
	//	}

	return &driver{db: db}, nil
}

// TODO fix transaction policy
func (d *driver) SaveTransaction(tx *models.Transaction) error {
	return d.db.Transaction(func(dbTx *gorm.DB) error {
		// Попытка вставки; при конфликте (например, уникальный hash) — вернуть ошибку
		if err := dbTx.Create(tx).Error; err != nil {
			// Можно дополнительно обработать ошибку конфликта, если нужно различать типы ошибок
			// Например, если используется PostgreSQL:
			// if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			//     return fmt.Errorf("transaction with same hash already exists")
			// }
			return err
		}
		return nil
	})
}

// TODO add offset
func (d *driver) GetTransactions(ctx context.Context, sender, receiver string) ([]models.Transaction, error) {
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
	if err := db.Find(&transactions).Limit(limit).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
