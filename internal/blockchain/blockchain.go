package blockchain

import (
    "errors"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type ERC20 struct {
    contract *bind.BoundContract
}

func NewERC20(address common.Address, client *ethclient.Client) (*ERC20, error) {
    abiJSON := `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`

    parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
    if err != nil {
        return nil, err
    }

    contract := bind.NewBoundContract(address, parsedABI, client, client, client)
    return &ERC20{contract: contract}, nil
}

func (e *ERC20) BalanceOf(opts *bind.CallOpts, address common.Address) (*big.Int, error) {
    var out []interface{}
    err := e.contract.Call(opts, &out, "balanceOf", address)
    if err != nil {
        return nil, err
    }

    if len(out) != 1 {
        return nil, errors.New("unexpected output from balanceOf")
    }

    balance, ok := out[0].(*big.Int)
    if !ok {
        return nil, errors.New("unexpected type in balanceOf output")
    }

    return balance, nil
}
