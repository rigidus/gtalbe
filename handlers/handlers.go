package handlers

import (
    "encoding/json"
    "net/http"
    "strings"

    "github.com/gorilla/mux"
    "go.uber.org/zap"
    "app/blockchain"
    "app/database"
    "app/models"
)

// Handler encapsulates dependencies for HTTP handlers.
type Handler struct {
    bc     blockchain.Client
    db     *gorm.DB
    logger *zap.Logger
}

// NewHandler creates a new Handler instance with dependencies.
func NewHandler(bc blockchain.Client, db *gorm.DB, logger *zap.Logger) *Handler {
    return &Handler{
        bc:     bc,
        db:     db,
        logger: logger,
    }
}

// GetBalance handles GET /balance/{address} to retrieve FIL and iFIL balances.
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    address := vars["address"]
    ctx := r.Context()

    // Валидация адреса
    if !strings.HasPrefix(address, "0x") || len(address) != 42 {
        h.logger.Warn("Invalid address format", zap.String("address", address))
        http.Error(w, "Invalid address format", http.StatusBadRequest)
        return
    }

    // Получение баланса $FIL
    filBalance, err := h.bc.GetFILBalance(ctx, address)
    if err != nil {
        h.logger.Error("Failed to get FIL balance", zap.String("address", address), zap.Error(err))
        http.Error(w, "Failed to get FIL balance", http.StatusInternalServerError)
        return
    }

    // Получение баланса iFIL
    ifilBalance, err := h.bc.GetIFILBalance(ctx, address)
    if err != nil {
        h.logger.Error("Failed to get iFIL balance", zap.String("address", address), zap.Error(err))
        http.Error(w, "Failed to get iFIL balance", http.StatusInternalServerError)
        return
    }

    // Формирование ответа
    response := map[string]string{
        "fil":  filBalance,
        "ifil": ifilBalance,
    }
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        h.logger.Error("Failed to encode response", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }

    h.logger.Info("Balance retrieved", zap.String("address", address), zap.String("fil", filBalance), zap.String("ifil", ifilBalance))
}

// SubmitTransaction handles POST /transaction to submit a new transaction.
func (h *Handler) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Декодирование запроса
    var req struct {
        SignedTx string  `json:"signedTx"`
        Sender   string  `json:"sender"`
        Receiver string  `json:"receiver"`
        Amount   float64 `json:"amount"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Warn("Invalid request body", zap.Error(err))
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Валидация входных данных
    if !strings.HasPrefix(req.SignedTx, "0x") || len(req.SignedTx) < 10 {
        h.logger.Warn("Invalid signed transaction format", zap.String("signedTx", req.SignedTx))
        http.Error(w, "Invalid signed transaction format", http.StatusBadRequest)
        return
    }
    if !strings.HasPrefix(req.Sender, "0x") || len(req.Sender) != 42 {
        h.logger.Warn("Invalid sender address", zap.String("sender", req.Sender))
        http.Error(w, "Invalid sender address", http.StatusBadRequest)
        return
    }
    if !strings.HasPrefix(req.Receiver, "0x") || len(req.Receiver) != 42 {
        h.logger.Warn("Invalid receiver address", zap.String("receiver", req.Receiver))
        http.Error(w, "Invalid receiver address", http.StatusBadRequest)
        return
    }
    if req.Amount <= 0 {
        h.logger.Warn("Invalid amount", zap.Float64("amount", req.Amount))
        http.Error(w, "Amount must be positive", http.StatusBadRequest)
        return
    }

    // Отправка транзакции в блокчейн
    txHash, err := h.bc.SubmitTransaction(ctx, req.SignedTx)
    if err != nil {
        h.logger.Error("Failed to submit transaction", zap.String("signedTx", req.SignedTx), zap.Error(err))
        http.Error(w, "Failed to submit transaction", http.StatusInternalServerError)
        return
    }

    // Сохранение транзакции в базе данных
    tx := &models.Transaction{
        Hash:     txHash,
        Sender:   req.Sender,
        Receiver: req.Receiver,
        Amount:   req.Amount,
        Status:   "pending",
    }
    if err := database.SaveTransaction(h.db, tx); err != nil {
        h.logger.Error("Failed to save transaction", zap.String("hash", txHash), zap.Error(err))
        http.Error(w, "Failed to save transaction", http.StatusInternalServerError)
        return
    }

    // Формирование ответа
    response := map[string]string{
        "hash": txHash,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(response); err != nil {
        h.logger.Error("Failed to encode response", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }

    h.logger.Info("Transaction submitted", zap.String("hash", txHash), zap.String("sender", req.Sender), zap.String("receiver", req.Receiver))
}

// GetTransactions handles GET /transactions to retrieve transaction records.
func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    query := r.URL.Query()

    // Получение фильтров из query-параметров
    sender := query.Get("sender")
    receiver := query.Get("receiver")

    // Валидация фильтров
    if sender != "" && (!strings.HasPrefix(sender, "0x") || len(sender) != 42) {
        h.logger.Warn("Invalid sender filter", zap.String("sender", sender))
        http.Error(w, "Invalid sender filter", http.StatusBadRequest)
        return
    }
    if receiver != "" && (!strings.HasPrefix(receiver, "0x") || len(receiver) != 42) {
        h.logger.Warn("Invalid receiver filter", zap.String("receiver", receiver))
        http.Error(w, "Invalid receiver filter", http.StatusBadRequest)
        return
    }

    // Построение запроса к базе данных
    dbQuery := h.db.WithContext(ctx)
    if sender != "" {
        dbQuery = dbQuery.Where("sender = ?", sender)
    }
    if receiver != "" {
        dbQuery = dbQuery.Where("receiver = ?", receiver)
    }

    // Извлечение транзакций
    var transactions []models.Transaction
    if err := dbQuery.Find(&transactions).Error; err != nil {
        h.logger.Error("Failed to retrieve transactions", zap.Error(err))
        http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
        return
    }

    // Формирование ответа
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(transactions); err != nil {
        h.logger.Error("Failed to encode response", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }

    h.logger.Info("Transactions retrieved", zap.Int("count", len(transactions)), zap.String("sender", sender), zap.String("receiver", receiver))
}
