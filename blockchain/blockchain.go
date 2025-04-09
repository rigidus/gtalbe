package blockchain

import (
    "context"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/glifio/go-pools-sdk/sdk" // Предполагаемый импорт для GLIF SDK
)

type BlockchainClient struct {
    glifClient  *sdk.Client
    ethClient   *ethclient.Client
    ifilAddress common.Address
}

func NewBlockchainClient(rpcURL, ifilContractAddress string) (*BlockchainClient, error) {
    // Инициализация GLIF клиента
    glifClient, err := sdk.NewClient(rpcURL)
    if err != nil {
        return nil, err
    }

    // Инициализация Ethereum клиента для iFIL
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }

    return &BlockchainClient{
        glifClient:  glifClient,
        ethClient:   ethClient,
        ifilAddress: common.HexToAddress(ifilContractAddress),
    }, nil
}

// GetFILBalance retrieves the $FIL balance for a given address.
func (c *BlockchainClient) GetFILBalance(ctx context.Context, address string) (string, error) {
    balance, err := c.glifClient.GetBalance(address)
    if err != nil {
        log.Printf("Failed to get FIL balance for %s: %v", address, err)
        return "", err
    }
    return balance.String(), nil // Предполагается, что balance — big.Int
}

// GetIFILBalance retrieves the iFIL balance for a given address.
func (c *BlockchainClient) GetIFILBalance(ctx context.Context, address string) (string, error) {
    // Предполагаем, что iFIL — стандартный ERC20 токен
    contract, err := NewERC20(c.ifilAddress, c.ethClient)
    if err != nil {
        return "", err
    }

    balance, err := contract.BalanceOf(&bind.CallOpts{Context: ctx}, common.HexToAddress(address))
    if err != nil {
        log.Printf("Failed to get iFIL balance for %s: %v", address, err)
        return "", err
    }
    return balance.String(), nil
}

// SubmitTransaction submits a signed transaction to the Filecoin network.
func (c *BlockchainClient) SubmitTransaction(ctx context.Context, signedTx string) (string, error) {
    txHash, err := c.glifClient.SubmitTransaction(signedTx)
    if err != nil {
        log.Printf("Failed to submit transaction: %v", err)
        return "", err
    }
    return txHash, nil
}

// ERC20 represents a minimal ERC20 contract interface (generated with abigen in real scenario).
type ERC20 struct {
    contract *bind.BoundContract
}

func NewERC20(address common.Address, client *ethclient.Client) (*ERC20, error) {
    // ABI для ERC20 метода balanceOf (в реальном проекте сгенерировать через abigen)
    abi := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`
    contract, err := bind.NewBoundContract(address, abi, client, client, client)
    if err != nil {
        return nil, err
    }
    return &ERC20{contract: contract}, nil
}

func (e *ERC20) BalanceOf(opts *bind.CallOpts, address common.Address) (*big.Int, error) {
    var result []*big.Int
    err := e.contract.Call(opts, &result, "balanceOf", address)
    if err != nil {
        return nil, err
    }
    return result[0], nil
}
