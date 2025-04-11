package server

import (
	"app/internal/blockchain"
	"app/internal/database/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"strings"

	"app/internal/database"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger

	e  *echo.Echo
	bc blockchain.Client
	db database.Database
}

func NewServer(bc blockchain.Client, db database.Database, logger *zap.Logger) *Server {
	e := echo.New()
	s := &Server{
		e:      e,
		bc:     bc,
		db:     db,
		logger: logger,
	}

	e.POST("/transaction", s.submitTransaction)
	e.GET("/transactions/", s.getTransactions)

	e.GET("/balance/:address", s.getBalance)
	return s
}

func (s *Server) Start(addr string) {
	go s.e.Start(addr)
	return
}

// GetBalance handles GET /balance/:address to retrieve FIL and iFIL balances.
func (s *Server) getBalance(c echo.Context) error {
	address := c.Param("address")
	ctx := c.Request().Context()

	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		s.logger.Warn("Invalid address format", zap.String("address", address))
		return ErrInvalidAddress
	}

	filBalance, err := s.bc.GetFILBalance(ctx, address)
	if err != nil {
		s.logger.Error("Failed to get FIL balance", zap.String("address", address), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "failed to get FIL balance"))
	}

	ifilBalance, err := s.bc.GetIFILBalance(ctx, address)
	if err != nil {
		s.logger.Error("Failed to get iFIL balance", zap.String("address", address), zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "failed to get iFIL balance"))
	}

	s.logger.Info("Balance retrieved", zap.String("address", address), zap.String("fil", filBalance), zap.String("ifil", ifilBalance))
	return c.JSON(http.StatusOK, &BalanceResponse{
		FIL:  filBalance,
		IFIL: ifilBalance,
	})
}

func (s *Server) submitTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	var req SubmitTransactionRequest
	if err := c.Bind(&req); err != nil {
		s.logger.Warn("Invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// TODO refactor errors
	if !strings.HasPrefix(req.SignedTx, "0x") || len(req.SignedTx) < 10 {
		s.logger.Warn("Invalid signed transaction format", zap.String("signedTx", req.SignedTx))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid signed transaction format"})
	}
	if !strings.HasPrefix(req.Sender, "0x") || len(req.Sender) != 42 {
		s.logger.Warn("Invalid sender address", zap.String("sender", req.Sender))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sender address"})
	}
	if !strings.HasPrefix(req.Receiver, "0x") || len(req.Receiver) != 42 {
		s.logger.Warn("Invalid receiver address", zap.String("receiver", req.Receiver))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receiver address"})
	}
	if req.Amount <= 0 {
		s.logger.Warn("Invalid amount", zap.Float64("amount", req.Amount))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Amount must be positive"})
	}

	txHash, err := s.bc.SubmitTransaction(ctx, req.SignedTx)
	if err != nil {
		s.logger.Error("Failed to submit transaction", zap.String("signedTx", req.SignedTx), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to submit transaction"})
	}

	tx := &models.Transaction{
		Hash:     txHash,
		Sender:   req.Sender,
		Receiver: req.Receiver,
		Amount:   req.Amount,
		Status:   "pending",
	}

	if err := s.db.SaveTransaction(tx); err != nil {
		s.logger.Error("Failed to save transaction", zap.String("hash", txHash), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save transaction"})
	}

	s.logger.Info("Transaction submitted", zap.String("hash", txHash), zap.String("sender", req.Sender), zap.String("receiver", req.Receiver))
	return c.JSON(http.StatusCreated, SubmitTransactionResponse{
		Hash: txHash,
	})
}

// GetTransactions handles GET /transactions to retrieve transaction records.
func (s *Server) getTransactions(c echo.Context) error {
	sender := c.QueryParam("sender")
	receiver := c.QueryParam("receiver")

	if sender != "" && (!strings.HasPrefix(sender, "0x") || len(sender) != 42) {
		s.logger.Warn("Invalid sender filter", zap.String("sender", sender))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sender filter"})
	}
	if receiver != "" && (!strings.HasPrefix(receiver, "0x") || len(receiver) != 42) {
		s.logger.Warn("Invalid receiver filter", zap.String("receiver", receiver))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receiver filter"})
	}

	txs, err := s.db.GetTransactions(c.Request().Context(), sender, receiver)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve transactions"})
	}

	s.logger.Info("Transactions retrieved", zap.Int("count", len(txs)), zap.String("sender", sender), zap.String("receiver", receiver))
	return c.JSON(http.StatusOK, txs)
}
