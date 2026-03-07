// Code generated — DO NOT EDIT.

//go:build !wasip1

package ioracle

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	evmmock "github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm/mock"
)

var (
	_ = errors.New
	_ = fmt.Errorf
	_ = big.NewInt
	_ = common.Big1
)

// IOracleMock is a mock implementation of IOracle for testing.
type IOracleMock struct {
	IfaPriceFeed               func() (common.Address, error)
	Owner                      func() (common.Address, error)
	OwnershipHandoverExpiresAt func(OwnershipHandoverExpiresAtInput) (*big.Int, error)
	RelayerNode                func() (common.Address, error)
}

// NewIOracleMock creates a new IOracleMock for testing.
func NewIOracleMock(address common.Address, clientMock *evmmock.ClientCapability) *IOracleMock {
	mock := &IOracleMock{}

	codec, err := NewCodec()
	if err != nil {
		panic("failed to create codec for mock: " + err.Error())
	}

	abi := codec.(*Codec).abi
	_ = abi

	funcMap := map[string]func([]byte) ([]byte, error){
		string(abi.Methods["IfaPriceFeed"].ID[:4]): func(payload []byte) ([]byte, error) {
			if mock.IfaPriceFeed == nil {
				return nil, errors.New("IfaPriceFeed method not mocked")
			}
			result, err := mock.IfaPriceFeed()
			if err != nil {
				return nil, err
			}
			return abi.Methods["IfaPriceFeed"].Outputs.Pack(result)
		},
		string(abi.Methods["owner"].ID[:4]): func(payload []byte) ([]byte, error) {
			if mock.Owner == nil {
				return nil, errors.New("owner method not mocked")
			}
			result, err := mock.Owner()
			if err != nil {
				return nil, err
			}
			return abi.Methods["owner"].Outputs.Pack(result)
		},
		string(abi.Methods["ownershipHandoverExpiresAt"].ID[:4]): func(payload []byte) ([]byte, error) {
			if mock.OwnershipHandoverExpiresAt == nil {
				return nil, errors.New("ownershipHandoverExpiresAt method not mocked")
			}
			inputs := abi.Methods["ownershipHandoverExpiresAt"].Inputs

			values, err := inputs.Unpack(payload)
			if err != nil {
				return nil, errors.New("Failed to unpack payload")
			}
			if len(values) != 1 {
				return nil, errors.New("expected 1 input value")
			}

			args := OwnershipHandoverExpiresAtInput{
				PendingOwner: values[0].(common.Address),
			}

			result, err := mock.OwnershipHandoverExpiresAt(args)
			if err != nil {
				return nil, err
			}
			return abi.Methods["ownershipHandoverExpiresAt"].Outputs.Pack(result)
		},
		string(abi.Methods["relayerNode"].ID[:4]): func(payload []byte) ([]byte, error) {
			if mock.RelayerNode == nil {
				return nil, errors.New("relayerNode method not mocked")
			}
			result, err := mock.RelayerNode()
			if err != nil {
				return nil, err
			}
			return abi.Methods["relayerNode"].Outputs.Pack(result)
		},
	}

	evmmock.AddContractMock(address, clientMock, funcMap, nil)
	return mock
}
