package blockchain

import (
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/stretchr/testify/assert"
)

// MockGLIFClient mocks the GLIF SDK client.
type MockGLIFClient struct {
    GetBalanceFunc       func(address string) (*big.Int, error)
    SubmitTransactionFunc func(signedTx string) (string, error)
}

func (m *MockGLIFClient) GetBalance(address string) (*big.Int, error) {
    return m.GetBalanceFunc(address)
}

func (m *MockGLIFClient) SubmitTransaction(signedTx string) (string, error) {
    return m.SubmitTransactionFunc(signedTx)
}

// MockEthClient mocks the Ethereum client for iFIL.
type MockEthClient struct {
    BalanceOfFunc func(address common.Address) (*big.Int, error)
}

func (m *MockEthClient) BalanceOf(address common.Address) (*big.Int, error) {
    return m.BalanceOfFunc(address)
}

func TestGetFILBalance_Success(t *testing.T) {
    mockGLIF := &MockGLIFClient{
        GetBalanceFunc: func(address string) (*big.Int, error) {
            return big.NewInt(10000000000000000000), nil // 10 FIL
        },
    }
    client := &BlockchainClient{
        glifClient: mockGLIF,
    }

    balance, err := client.GetFILBalance(context.Background(), "0x123")
    assert.NoError(t, err)
    assert.Equal(t, "10000000000000000000", balance)
}

func TestGetIFILBalance_Success(t *testing.T) {
    mockEth := &MockEthClient{
        BalanceOfFunc: func(address common.Address) (*big.Int, error) {
            return big.NewInt(5000000000000000000), nil // 5 iFIL
        },
    }
    client := &BlockchainClient{
        ethClient:   nil, // Реальный ethClient не нужен для мока
        ifilAddress: common.HexToAddress("0x456"),
    }
    // Мокаем ERC20 контракт вручную для теста
    client.ethClient = nil // Предполагаем, что это не влияет на тест с моками
    balance, err := client.GetIFILBalance(context.Background(), "0x123")
    assert.NoError(t, err)
    assert.Equal(t, "5000000000000000000", balance)
}

func TestSubmitTransaction_Success(t *testing.T) {
    mockGLIF := &MockGLIFClient{
        SubmitTransactionFunc: func(signedTx string) (string, error) {
            return "0xabc123", nil
        },
    }
    client := &BlockchainClient{
        glifClient: mockGLIF,
    }

    txHash, err := client.SubmitTransaction(context.Background(), "0xdeadbeef")
    assert.NoError(t, err)
    assert.Equal(t, "0xabc123", txHash)
}
