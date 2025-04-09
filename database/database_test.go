package database

import (
    "context"
    "log"
    "os"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
    ctx := context.Background()
    postgresContainer, err := postgres.Run(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer postgresContainer.Terminate(ctx)

    connectionString, err := postgresContainer.ConnectionString(ctx)
    if err != nil {
        log.Fatal(err)
    }

    db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    db.AutoMigrate(&models.Transaction{})

    testDB = db

    code := m.Run()

    os.Exit(code)
}


func TestSaveTransaction_Success(t *testing.T) {
    tx := &models.Transaction{
        Hash:     "0xabc",
        Sender:   "0x123",
        Receiver: "0x456",
        Amount:   100,
        Status:   "pending",
    }

    err := SaveTransaction(testDB, tx)
    if err != nil {
        t.Errorf("failed to save transaction: %v", err)
    }

    var savedTx models.Transaction
    testDB.First(&savedTx, "hash = ?", "0xabc")
    if savedTx.ID == 0 {
        t.Error("transaction not saved")
    }
}

func TestGetTransactions_Success(t *testing.T) {
    // Предварительно сохраняем транзакцию для теста
    tx := &models.Transaction{
        Hash:     "0xabc",
        Sender:   "0x123",
        Receiver: "0x456",
        Amount:   100,
        Status:   "pending",
    }
    testDB.Create(tx)

    transactions := GetTransactions(testDB)
    if len(transactions) == 0 {
        t.Error("expected transactions but got none")
    }
}
