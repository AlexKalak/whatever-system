package service

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func callString(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, contract common.Address, method string) (string, error) {
	data, err := contractABI.Pack(method)
	if err != nil {
		return "", err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &contract, Data: data}, nil)
	if err != nil {
		return "", err
	}
	vals, err := contractABI.Unpack(method, out)
	if err != nil || len(vals) == 0 {
		return "", fmt.Errorf("unpack %s failed: %w", method, err)
	}
	s, ok := vals[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected %s return type", method)
	}
	return s, nil
}

func callUint8(ctx context.Context, client *ethclient.Client, contractABI abi.ABI, contract common.Address, method string) (uint8, error) {
	data, err := contractABI.Pack(method)
	if err != nil {
		return 0, err
	}
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &contract, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	vals, err := contractABI.Unpack(method, out)
	if err != nil || len(vals) == 0 {
		return 0, fmt.Errorf("unpack %s failed: %w", method, err)
	}
	d, ok := vals[0].(uint8)
	if !ok {
		return 0, fmt.Errorf("unexpected %s return type", method)
	}
	return d, nil
}
