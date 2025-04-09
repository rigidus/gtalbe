package database

import "gorm.io/gorm"

func SaveTransaction(db *gorm.DB, tx *models.Transaction) error {
    return db.Create(tx).Error
}

func GetTransactions(db *gorm.DB) []models.Transaction {
    var transactions []models.Transaction
    db.Find(&transactions)
    return transactions
}
