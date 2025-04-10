package database

import (
    "gorm.io/gorm"
    "app/models" // ← имя модуля из go.mod + путь к папке
)

func SaveTransaction(db *gorm.DB, tx *models.Transaction) error {
    return db.Create(tx).Error
}

func GetTransactions(db *gorm.DB) []models.Transaction {
    var transactions []models.Transaction
    db.Find(&transactions)
    return transactions
}
