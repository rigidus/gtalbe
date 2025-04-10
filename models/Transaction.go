package models

import (
    "time"
)

type Transaction struct {
    ID        uint64    `gorm:"primaryKey"`
    Hash      string    `gorm:"type:varchar(66);unique"`
    Sender    string    `gorm:"type:varchar(42)"`
    Receiver  string    `gorm:"type:varchar(42)"`
    Amount    float64   `gorm:"type:decimal(30,18)"`
    Timestamp time.Time `gorm:"type:timestamp with time zone"`
    Status    string    `gorm:"type:enum('pending', 'confirmed', 'failed');default:'pending'"`
}
