package blockchain

import "context"

// Client defines the interface for blockchain interactions.
type Client interface {
    GetFILBalance(ctx context.Context, address string) (string, error)
    GetIFILBalance(ctx context.Context, address string) (string, error)
    SubmitTransaction(ctx context.Context, signedTx string) (string, error)
}
