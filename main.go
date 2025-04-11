package main

import (
	"app/internal/blockchain"
	"app/internal/database"
	"app/internal/server"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
	if rpcURL == "" {
		logger.Fatal("BLOCKCHAIN_RPC_URL environment variable not set")
	}

	ifilAddress := os.Getenv("IFIL_CONTRACT_ADDRESS_HEX")
	if ifilAddress == "" {
		logger.Fatal("IFIL_CONTRACT_ADDRESS environment variable not set")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		logger.Fatal("Failed to connect to ethereum client", zap.Error(err))
	}

	// TODO check if address is valid
	_, err = blockchain.NewERC20(common.HexToAddress(ifilAddress), client)
	if err != nil {
		logger.Fatal("Failed to initialize blockchain client", zap.Error(err))
	}

	dbURL := os.Getenv("DATABASE_DNS")
	dbDriver, err := database.NewDriver(dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// TODO pass realy blockchain client
	srv := server.NewServer(nil, dbDriver, logger)
	srv.Start(":8080")

	ctx := context.Background()
	<-ctx.Done()
}
