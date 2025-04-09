package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "go.uber.org/zap"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/yourproject/blockchain"
    "github.com/yourproject/handlers"
)

func main() {
    // Инициализация логгера
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Подключение к базе данных
    dbURL := os.Getenv("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }

    // Инициализация клиента блокчейна
    rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
    ifilAddress := os.Getenv("IFIL_CONTRACT_ADDRESS")
    bc, err := blockchain.NewBlockchainClient(rpcURL, ifilAddress)
    if err != nil {
        logger.Fatal("Failed to initialize blockchain client", zap.Error(err))
    }

    // Инициализация обработчиков
    h := handlers.NewHandler(bc, db, logger)

    // Настройка маршрутов
    r := mux.NewRouter()
    r.HandleFunc("/balance/{address}", h.GetBalance).Methods("GET")
    r.HandleFunc("/transaction", h.SubmitTransaction).Methods("POST")
    r.HandleFunc("/transactions", h.GetTransactions).Methods("GET")

    // Запуск сервера
    port := ":8080"
    logger.Info("Starting server", zap.String("port", port))
    if err := http.ListenAndServe(port, r); err != nil {
        logger.Fatal("Server failed", zap.Error(err))
    }
}
