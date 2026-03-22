// Code generated — DO NOT EDIT.

package ioracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb2 "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm/bindings"
	"github.com/smartcontractkit/cre-sdk-go/cre"
)

var (
	_ = bytes.Equal
	_ = errors.New
	_ = fmt.Sprintf
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
	_ = emptypb.Empty{}
	_ = pb.NewBigIntFromInt
	_ = pb2.AggregationType_AGGREGATION_TYPE_COMMON_PREFIX
	_ = bindings.FilterOptions{}
	_ = evm.FilterLogTriggerRequest{}
	_ = cre.ResponseBufferTooSmall
	_ = rpc.API{}
	_ = json.Unmarshal
	_ = reflect.Bool
)

var IOracleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_relayerNode\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_IIfaPriceFeed\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_forwarderAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"IfaPriceFeed\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIIfaPriceFeed\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cancelOwnershipHandover\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"completeOwnershipHandover\",\"inputs\":[{\"name\":\"pendingOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"getExpectedAuthor\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExpectedWorkflowId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getExpectedWorkflowName\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes10\",\"internalType\":\"bytes10\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getForwarderAddress\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"onReport\",\"inputs\":[{\"name\":\"metadata\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"report\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"result\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ownershipHandoverExpiresAt\",\"inputs\":[{\"name\":\"pendingOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"result\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proceesTheSumission\",\"inputs\":[{\"name\":\"data\",\"type\":\"tuple\",\"internalType\":\"structIfaPriceFeedVerifier.SumissionData\",\"components\":[{\"name\":\"assesetindex\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"price\",\"type\":\"tuple\",\"internalType\":\"structIIfaPriceFeed.PriceFeed\",\"components\":[{\"name\":\"price\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"decimal\",\"type\":\"int8\",\"internalType\":\"int8\"},{\"name\":\"lastUpdateTime\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"relayerNode\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"requestOwnershipHandover\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"setExpectedAuthor\",\"inputs\":[{\"name\":\"_author\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setExpectedWorkflowId\",\"inputs\":[{\"name\":\"_id\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setExpectedWorkflowName\",\"inputs\":[{\"name\":\"_name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setForwarderAddress\",\"inputs\":[{\"name\":\"_forwarder\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRelayerNode\",\"inputs\":[{\"name\":\"_relayerNode\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitPriceFeed\",\"inputs\":[{\"name\":\"_assetindex\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"_prices\",\"type\":\"tuple[]\",\"internalType\":\"structIIfaPriceFeed.PriceFeed[]\",\"components\":[{\"name\":\"price\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"decimal\",\"type\":\"int8\",\"internalType\":\"int8\"},{\"name\":\"lastUpdateTime\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"event\",\"name\":\"ExpectedAuthorUpdated\",\"inputs\":[{\"name\":\"previousAuthor\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAuthor\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ExpectedWorkflowIdUpdated\",\"inputs\":[{\"name\":\"previousId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ExpectedWorkflowNameUpdated\",\"inputs\":[{\"name\":\"previousName\",\"type\":\"bytes10\",\"indexed\":true,\"internalType\":\"bytes10\"},{\"name\":\"newName\",\"type\":\"bytes10\",\"indexed\":true,\"internalType\":\"bytes10\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ForwarderAddressUpdated\",\"inputs\":[{\"name\":\"previousForwarder\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newForwarder\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipHandoverCanceled\",\"inputs\":[{\"name\":\"pendingOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipHandoverRequested\",\"inputs\":[{\"name\":\"pendingOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"oldOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RelayerNodeSet\",\"inputs\":[{\"name\":\"newRelayerNode\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldRelayerNode\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SecurityWarning\",\"inputs\":[{\"name\":\"message\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AlreadyInitialized\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAssePrice\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAssetIndexorPriceLength\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAuthor\",\"inputs\":[{\"name\":\"received\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expected\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidForwarderAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRelayerNode\",\"inputs\":[{\"name\":\"_address\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidSender\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"expected\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidWorkflowId\",\"inputs\":[{\"name\":\"received\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expected\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidWorkflowName\",\"inputs\":[{\"name\":\"received\",\"type\":\"bytes10\",\"internalType\":\"bytes10\"},{\"name\":\"expected\",\"type\":\"bytes10\",\"internalType\":\"bytes10\"}]},{\"type\":\"error\",\"name\":\"NewOwnerIsZeroAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoHandoverRequest\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OnlyRelayerNode\",\"inputs\":[{\"name\":\"_caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"Unauthorized\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"WorkflowNameRequiresAuthorValidation\",\"inputs\":[]}]",
}

// Structs
type IIfaPriceFeedPriceFeed struct {
	Price          *big.Int
	Decimal        int8
	LastUpdateTime uint64
}

type IfaPriceFeedVerifierSumissionData struct {
	Assesetindex [32]byte
	Price        IIfaPriceFeedPriceFeed
}

// Contract Method Inputs
type CompleteOwnershipHandoverInput struct {
	PendingOwner common.Address
}

type OnReportInput struct {
	Metadata []byte
	Report   []byte
}

type OwnershipHandoverExpiresAtInput struct {
	PendingOwner common.Address
}

type ProceesTheSumissionInput struct {
	Data IfaPriceFeedVerifierSumissionData
}

type SetExpectedAuthorInput struct {
	Author common.Address
}

type SetExpectedWorkflowIdInput struct {
	Id [32]byte
}

type SetExpectedWorkflowNameInput struct {
	Name string
}

type SetForwarderAddressInput struct {
	Forwarder common.Address
}

type SetRelayerNodeInput struct {
	RelayerNode common.Address
}

type SubmitPriceFeedInput struct {
	Assetindex [][32]byte
	Prices     []IIfaPriceFeedPriceFeed
}

type SupportsInterfaceInput struct {
	InterfaceId [4]byte
}

type TransferOwnershipInput struct {
	NewOwner common.Address
}

// Contract Method Outputs

// Errors
type AlreadyInitialized struct {
}

type InvalidAssePrice struct {
}

type InvalidAssetIndexorPriceLength struct {
}

type InvalidAuthor struct {
	Received common.Address
	Expected common.Address
}

type InvalidForwarderAddress struct {
}

type InvalidRelayerNode struct {
	Address common.Address
}

type InvalidSender struct {
	Sender   common.Address
	Expected common.Address
}

type InvalidWorkflowId struct {
	Received [32]byte
	Expected [32]byte
}

type InvalidWorkflowName struct {
	Received [10]byte
	Expected [10]byte
}

type NewOwnerIsZeroAddress struct {
}

type NoHandoverRequest struct {
}

type OnlyRelayerNode struct {
	Caller common.Address
}

type Unauthorized struct {
}

type WorkflowNameRequiresAuthorValidation struct {
}

// Events
// The <Event>Topics struct should be used as a filter (for log triggers).
// Note: It is only possible to filter on indexed fields.
// Indexed (string and bytes) fields will be of type common.Hash.
// They need to he (crypto.Keccak256) hashed and passed in.
// Indexed (tuple/slice/array) fields can be passed in as is, the Encode<Event>Topics function will handle the hashing.
//
// The <Event>Decoded struct will be the result of calling decode (Adapt) on the log trigger result.
// Indexed dynamic type fields will be of type common.Hash.

type ExpectedAuthorUpdatedTopics struct {
	PreviousAuthor common.Address
	NewAuthor      common.Address
}

type ExpectedAuthorUpdatedDecoded struct {
	PreviousAuthor common.Address
	NewAuthor      common.Address
}

type ExpectedWorkflowIdUpdatedTopics struct {
	PreviousId [32]byte
	NewId      [32]byte
}

type ExpectedWorkflowIdUpdatedDecoded struct {
	PreviousId [32]byte
	NewId      [32]byte
}

type ExpectedWorkflowNameUpdatedTopics struct {
	PreviousName [10]byte
	NewName      [10]byte
}

type ExpectedWorkflowNameUpdatedDecoded struct {
	PreviousName [10]byte
	NewName      [10]byte
}

type ForwarderAddressUpdatedTopics struct {
	PreviousForwarder common.Address
	NewForwarder      common.Address
}

type ForwarderAddressUpdatedDecoded struct {
	PreviousForwarder common.Address
	NewForwarder      common.Address
}

type OwnershipHandoverCanceledTopics struct {
	PendingOwner common.Address
}

type OwnershipHandoverCanceledDecoded struct {
	PendingOwner common.Address
}

type OwnershipHandoverRequestedTopics struct {
	PendingOwner common.Address
}

type OwnershipHandoverRequestedDecoded struct {
	PendingOwner common.Address
}

type OwnershipTransferredTopics struct {
	OldOwner common.Address
	NewOwner common.Address
}

type OwnershipTransferredDecoded struct {
	OldOwner common.Address
	NewOwner common.Address
}

type RelayerNodeSetTopics struct {
	NewRelayerNode common.Address
	OldRelayerNode common.Address
}

type RelayerNodeSetDecoded struct {
	NewRelayerNode common.Address
	OldRelayerNode common.Address
}

type SecurityWarningTopics struct {
}

type SecurityWarningDecoded struct {
	Message string
}

// Main Binding Type for IOracle
type IOracle struct {
	Address common.Address
	Options *bindings.ContractInitOptions
	ABI     *abi.ABI
	client  *evm.Client
	Codec   IOracleCodec
}

type IOracleCodec interface {
	EncodeIfaPriceFeedMethodCall() ([]byte, error)
	DecodeIfaPriceFeedMethodOutput(data []byte) (common.Address, error)
	EncodeCancelOwnershipHandoverMethodCall() ([]byte, error)
	EncodeCompleteOwnershipHandoverMethodCall(in CompleteOwnershipHandoverInput) ([]byte, error)
	EncodeGetExpectedAuthorMethodCall() ([]byte, error)
	DecodeGetExpectedAuthorMethodOutput(data []byte) (common.Address, error)
	EncodeGetExpectedWorkflowIdMethodCall() ([]byte, error)
	DecodeGetExpectedWorkflowIdMethodOutput(data []byte) ([32]byte, error)
	EncodeGetExpectedWorkflowNameMethodCall() ([]byte, error)
	DecodeGetExpectedWorkflowNameMethodOutput(data []byte) ([10]byte, error)
	EncodeGetForwarderAddressMethodCall() ([]byte, error)
	DecodeGetForwarderAddressMethodOutput(data []byte) (common.Address, error)
	EncodeOnReportMethodCall(in OnReportInput) ([]byte, error)
	EncodeOwnerMethodCall() ([]byte, error)
	DecodeOwnerMethodOutput(data []byte) (common.Address, error)
	EncodeOwnershipHandoverExpiresAtMethodCall(in OwnershipHandoverExpiresAtInput) ([]byte, error)
	DecodeOwnershipHandoverExpiresAtMethodOutput(data []byte) (*big.Int, error)
	EncodeProceesTheSumissionMethodCall(in ProceesTheSumissionInput) ([]byte, error)
	EncodeRelayerNodeMethodCall() ([]byte, error)
	DecodeRelayerNodeMethodOutput(data []byte) (common.Address, error)
	EncodeRenounceOwnershipMethodCall() ([]byte, error)
	EncodeRequestOwnershipHandoverMethodCall() ([]byte, error)
	EncodeSetExpectedAuthorMethodCall(in SetExpectedAuthorInput) ([]byte, error)
	EncodeSetExpectedWorkflowIdMethodCall(in SetExpectedWorkflowIdInput) ([]byte, error)
	EncodeSetExpectedWorkflowNameMethodCall(in SetExpectedWorkflowNameInput) ([]byte, error)
	EncodeSetForwarderAddressMethodCall(in SetForwarderAddressInput) ([]byte, error)
	EncodeSetRelayerNodeMethodCall(in SetRelayerNodeInput) ([]byte, error)
	EncodeSubmitPriceFeedMethodCall(in SubmitPriceFeedInput) ([]byte, error)
	EncodeSupportsInterfaceMethodCall(in SupportsInterfaceInput) ([]byte, error)
	DecodeSupportsInterfaceMethodOutput(data []byte) (bool, error)
	EncodeTransferOwnershipMethodCall(in TransferOwnershipInput) ([]byte, error)
	EncodeIIfaPriceFeedPriceFeedStruct(in IIfaPriceFeedPriceFeed) ([]byte, error)
	EncodeIfaPriceFeedVerifierSumissionDataStruct(in IfaPriceFeedVerifierSumissionData) ([]byte, error)
	ExpectedAuthorUpdatedLogHash() []byte
	EncodeExpectedAuthorUpdatedTopics(evt abi.Event, values []ExpectedAuthorUpdatedTopics) ([]*evm.TopicValues, error)
	DecodeExpectedAuthorUpdated(log *evm.Log) (*ExpectedAuthorUpdatedDecoded, error)
	ExpectedWorkflowIdUpdatedLogHash() []byte
	EncodeExpectedWorkflowIdUpdatedTopics(evt abi.Event, values []ExpectedWorkflowIdUpdatedTopics) ([]*evm.TopicValues, error)
	DecodeExpectedWorkflowIdUpdated(log *evm.Log) (*ExpectedWorkflowIdUpdatedDecoded, error)
	ExpectedWorkflowNameUpdatedLogHash() []byte
	EncodeExpectedWorkflowNameUpdatedTopics(evt abi.Event, values []ExpectedWorkflowNameUpdatedTopics) ([]*evm.TopicValues, error)
	DecodeExpectedWorkflowNameUpdated(log *evm.Log) (*ExpectedWorkflowNameUpdatedDecoded, error)
	ForwarderAddressUpdatedLogHash() []byte
	EncodeForwarderAddressUpdatedTopics(evt abi.Event, values []ForwarderAddressUpdatedTopics) ([]*evm.TopicValues, error)
	DecodeForwarderAddressUpdated(log *evm.Log) (*ForwarderAddressUpdatedDecoded, error)
	OwnershipHandoverCanceledLogHash() []byte
	EncodeOwnershipHandoverCanceledTopics(evt abi.Event, values []OwnershipHandoverCanceledTopics) ([]*evm.TopicValues, error)
	DecodeOwnershipHandoverCanceled(log *evm.Log) (*OwnershipHandoverCanceledDecoded, error)
	OwnershipHandoverRequestedLogHash() []byte
	EncodeOwnershipHandoverRequestedTopics(evt abi.Event, values []OwnershipHandoverRequestedTopics) ([]*evm.TopicValues, error)
	DecodeOwnershipHandoverRequested(log *evm.Log) (*OwnershipHandoverRequestedDecoded, error)
	OwnershipTransferredLogHash() []byte
	EncodeOwnershipTransferredTopics(evt abi.Event, values []OwnershipTransferredTopics) ([]*evm.TopicValues, error)
	DecodeOwnershipTransferred(log *evm.Log) (*OwnershipTransferredDecoded, error)
	RelayerNodeSetLogHash() []byte
	EncodeRelayerNodeSetTopics(evt abi.Event, values []RelayerNodeSetTopics) ([]*evm.TopicValues, error)
	DecodeRelayerNodeSet(log *evm.Log) (*RelayerNodeSetDecoded, error)
	SecurityWarningLogHash() []byte
	EncodeSecurityWarningTopics(evt abi.Event, values []SecurityWarningTopics) ([]*evm.TopicValues, error)
	DecodeSecurityWarning(log *evm.Log) (*SecurityWarningDecoded, error)
}

func NewIOracle(
	client *evm.Client,
	address common.Address,
	options *bindings.ContractInitOptions,
) (*IOracle, error) {
	parsed, err := abi.JSON(strings.NewReader(IOracleMetaData.ABI))
	if err != nil {
		return nil, err
	}
	codec, err := NewCodec()
	if err != nil {
		return nil, err
	}
	return &IOracle{
		Address: address,
		Options: options,
		ABI:     &parsed,
		client:  client,
		Codec:   codec,
	}, nil
}

type Codec struct {
	abi *abi.ABI
}

func NewCodec() (IOracleCodec, error) {
	parsed, err := abi.JSON(strings.NewReader(IOracleMetaData.ABI))
	if err != nil {
		return nil, err
	}
	return &Codec{abi: &parsed}, nil
}

func (c *Codec) EncodeIfaPriceFeedMethodCall() ([]byte, error) {
	return c.abi.Pack("IfaPriceFeed")
}

func (c *Codec) DecodeIfaPriceFeedMethodOutput(data []byte) (common.Address, error) {
	vals, err := c.abi.Methods["IfaPriceFeed"].Outputs.Unpack(data)
	if err != nil {
		return *new(common.Address), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(common.Address), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result common.Address
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(common.Address), fmt.Errorf("failed to unmarshal to common.Address: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeCancelOwnershipHandoverMethodCall() ([]byte, error) {
	return c.abi.Pack("cancelOwnershipHandover")
}

func (c *Codec) EncodeCompleteOwnershipHandoverMethodCall(in CompleteOwnershipHandoverInput) ([]byte, error) {
	return c.abi.Pack("completeOwnershipHandover", in.PendingOwner)
}

func (c *Codec) EncodeGetExpectedAuthorMethodCall() ([]byte, error) {
	return c.abi.Pack("getExpectedAuthor")
}

func (c *Codec) DecodeGetExpectedAuthorMethodOutput(data []byte) (common.Address, error) {
	vals, err := c.abi.Methods["getExpectedAuthor"].Outputs.Unpack(data)
	if err != nil {
		return *new(common.Address), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(common.Address), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result common.Address
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(common.Address), fmt.Errorf("failed to unmarshal to common.Address: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeGetExpectedWorkflowIdMethodCall() ([]byte, error) {
	return c.abi.Pack("getExpectedWorkflowId")
}

func (c *Codec) DecodeGetExpectedWorkflowIdMethodOutput(data []byte) ([32]byte, error) {
	vals, err := c.abi.Methods["getExpectedWorkflowId"].Outputs.Unpack(data)
	if err != nil {
		return *new([32]byte), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new([32]byte), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result [32]byte
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new([32]byte), fmt.Errorf("failed to unmarshal to [32]byte: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeGetExpectedWorkflowNameMethodCall() ([]byte, error) {
	return c.abi.Pack("getExpectedWorkflowName")
}

func (c *Codec) DecodeGetExpectedWorkflowNameMethodOutput(data []byte) ([10]byte, error) {
	vals, err := c.abi.Methods["getExpectedWorkflowName"].Outputs.Unpack(data)
	if err != nil {
		return *new([10]byte), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new([10]byte), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result [10]byte
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new([10]byte), fmt.Errorf("failed to unmarshal to [10]byte: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeGetForwarderAddressMethodCall() ([]byte, error) {
	return c.abi.Pack("getForwarderAddress")
}

func (c *Codec) DecodeGetForwarderAddressMethodOutput(data []byte) (common.Address, error) {
	vals, err := c.abi.Methods["getForwarderAddress"].Outputs.Unpack(data)
	if err != nil {
		return *new(common.Address), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(common.Address), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result common.Address
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(common.Address), fmt.Errorf("failed to unmarshal to common.Address: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeOnReportMethodCall(in OnReportInput) ([]byte, error) {
	return c.abi.Pack("onReport", in.Metadata, in.Report)
}

func (c *Codec) EncodeOwnerMethodCall() ([]byte, error) {
	return c.abi.Pack("owner")
}

func (c *Codec) DecodeOwnerMethodOutput(data []byte) (common.Address, error) {
	vals, err := c.abi.Methods["owner"].Outputs.Unpack(data)
	if err != nil {
		return *new(common.Address), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(common.Address), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result common.Address
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(common.Address), fmt.Errorf("failed to unmarshal to common.Address: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeOwnershipHandoverExpiresAtMethodCall(in OwnershipHandoverExpiresAtInput) ([]byte, error) {
	return c.abi.Pack("ownershipHandoverExpiresAt", in.PendingOwner)
}

func (c *Codec) DecodeOwnershipHandoverExpiresAtMethodOutput(data []byte) (*big.Int, error) {
	vals, err := c.abi.Methods["ownershipHandoverExpiresAt"].Outputs.Unpack(data)
	if err != nil {
		return *new(*big.Int), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(*big.Int), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result *big.Int
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(*big.Int), fmt.Errorf("failed to unmarshal to *big.Int: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeProceesTheSumissionMethodCall(in ProceesTheSumissionInput) ([]byte, error) {
	return c.abi.Pack("proceesTheSumission", in.Data)
}

func (c *Codec) EncodeRelayerNodeMethodCall() ([]byte, error) {
	return c.abi.Pack("relayerNode")
}

func (c *Codec) DecodeRelayerNodeMethodOutput(data []byte) (common.Address, error) {
	vals, err := c.abi.Methods["relayerNode"].Outputs.Unpack(data)
	if err != nil {
		return *new(common.Address), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(common.Address), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result common.Address
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(common.Address), fmt.Errorf("failed to unmarshal to common.Address: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeRenounceOwnershipMethodCall() ([]byte, error) {
	return c.abi.Pack("renounceOwnership")
}

func (c *Codec) EncodeRequestOwnershipHandoverMethodCall() ([]byte, error) {
	return c.abi.Pack("requestOwnershipHandover")
}

func (c *Codec) EncodeSetExpectedAuthorMethodCall(in SetExpectedAuthorInput) ([]byte, error) {
	return c.abi.Pack("setExpectedAuthor", in.Author)
}

func (c *Codec) EncodeSetExpectedWorkflowIdMethodCall(in SetExpectedWorkflowIdInput) ([]byte, error) {
	return c.abi.Pack("setExpectedWorkflowId", in.Id)
}

func (c *Codec) EncodeSetExpectedWorkflowNameMethodCall(in SetExpectedWorkflowNameInput) ([]byte, error) {
	return c.abi.Pack("setExpectedWorkflowName", in.Name)
}

func (c *Codec) EncodeSetForwarderAddressMethodCall(in SetForwarderAddressInput) ([]byte, error) {
	return c.abi.Pack("setForwarderAddress", in.Forwarder)
}

func (c *Codec) EncodeSetRelayerNodeMethodCall(in SetRelayerNodeInput) ([]byte, error) {
	return c.abi.Pack("setRelayerNode", in.RelayerNode)
}

func (c *Codec) EncodeSubmitPriceFeedMethodCall(in SubmitPriceFeedInput) ([]byte, error) {
	return c.abi.Pack("submitPriceFeed", in.Assetindex, in.Prices)
}

func (c *Codec) EncodeSupportsInterfaceMethodCall(in SupportsInterfaceInput) ([]byte, error) {
	return c.abi.Pack("supportsInterface", in.InterfaceId)
}

func (c *Codec) DecodeSupportsInterfaceMethodOutput(data []byte) (bool, error) {
	vals, err := c.abi.Methods["supportsInterface"].Outputs.Unpack(data)
	if err != nil {
		return *new(bool), err
	}
	jsonData, err := json.Marshal(vals[0])
	if err != nil {
		return *new(bool), fmt.Errorf("failed to marshal ABI result: %w", err)
	}

	var result bool
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return *new(bool), fmt.Errorf("failed to unmarshal to bool: %w", err)
	}

	return result, nil
}

func (c *Codec) EncodeTransferOwnershipMethodCall(in TransferOwnershipInput) ([]byte, error) {
	return c.abi.Pack("transferOwnership", in.NewOwner)
}

func (c *Codec) EncodeIIfaPriceFeedPriceFeedStruct(in IIfaPriceFeedPriceFeed) ([]byte, error) {
	tupleType, err := abi.NewType(
		"tuple", "",
		[]abi.ArgumentMarshaling{
			{Name: "price", Type: "int256"},
			{Name: "decimal", Type: "int8"},
			{Name: "lastUpdateTime", Type: "uint64"},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tuple type for IIfaPriceFeedPriceFeed: %w", err)
	}
	args := abi.Arguments{
		{Name: "iIfaPriceFeedPriceFeed", Type: tupleType},
	}

	return args.Pack(in)
}
func (c *Codec) EncodeIfaPriceFeedVerifierSumissionDataStruct(in IfaPriceFeedVerifierSumissionData) ([]byte, error) {
	tupleType, err := abi.NewType(
		"tuple", "",
		[]abi.ArgumentMarshaling{
			{Name: "assesetindex", Type: "bytes32"},
			{Name: "price", Type: "tuple", Components: []abi.ArgumentMarshaling{
				{Name: "price", Type: "int256"},
				{Name: "decimal", Type: "int8"},
				{Name: "lastUpdateTime", Type: "uint64"},
			}},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tuple type for IfaPriceFeedVerifierSumissionData: %w", err)
	}
	args := abi.Arguments{
		{Name: "ifaPriceFeedVerifierSumissionData", Type: tupleType},
	}

	return args.Pack(in)
}

func (c *Codec) ExpectedAuthorUpdatedLogHash() []byte {
	return c.abi.Events["ExpectedAuthorUpdated"].ID.Bytes()
}

func (c *Codec) EncodeExpectedAuthorUpdatedTopics(
	evt abi.Event,
	values []ExpectedAuthorUpdatedTopics,
) ([]*evm.TopicValues, error) {
	var previousAuthorRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.PreviousAuthor).IsZero() {
			previousAuthorRule = append(previousAuthorRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.PreviousAuthor)
		if err != nil {
			return nil, err
		}
		previousAuthorRule = append(previousAuthorRule, fieldVal)
	}
	var newAuthorRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.NewAuthor).IsZero() {
			newAuthorRule = append(newAuthorRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[1], v.NewAuthor)
		if err != nil {
			return nil, err
		}
		newAuthorRule = append(newAuthorRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		previousAuthorRule,
		newAuthorRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeExpectedAuthorUpdated decodes a log into a ExpectedAuthorUpdated struct.
func (c *Codec) DecodeExpectedAuthorUpdated(log *evm.Log) (*ExpectedAuthorUpdatedDecoded, error) {
	event := new(ExpectedAuthorUpdatedDecoded)
	if err := c.abi.UnpackIntoInterface(event, "ExpectedAuthorUpdated", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["ExpectedAuthorUpdated"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) ExpectedWorkflowIdUpdatedLogHash() []byte {
	return c.abi.Events["ExpectedWorkflowIdUpdated"].ID.Bytes()
}

func (c *Codec) EncodeExpectedWorkflowIdUpdatedTopics(
	evt abi.Event,
	values []ExpectedWorkflowIdUpdatedTopics,
) ([]*evm.TopicValues, error) {
	var previousIdRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.PreviousId).IsZero() {
			previousIdRule = append(previousIdRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.PreviousId)
		if err != nil {
			return nil, err
		}
		previousIdRule = append(previousIdRule, fieldVal)
	}
	var newIdRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.NewId).IsZero() {
			newIdRule = append(newIdRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[1], v.NewId)
		if err != nil {
			return nil, err
		}
		newIdRule = append(newIdRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		previousIdRule,
		newIdRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeExpectedWorkflowIdUpdated decodes a log into a ExpectedWorkflowIdUpdated struct.
func (c *Codec) DecodeExpectedWorkflowIdUpdated(log *evm.Log) (*ExpectedWorkflowIdUpdatedDecoded, error) {
	event := new(ExpectedWorkflowIdUpdatedDecoded)
	if err := c.abi.UnpackIntoInterface(event, "ExpectedWorkflowIdUpdated", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["ExpectedWorkflowIdUpdated"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) ExpectedWorkflowNameUpdatedLogHash() []byte {
	return c.abi.Events["ExpectedWorkflowNameUpdated"].ID.Bytes()
}

func (c *Codec) EncodeExpectedWorkflowNameUpdatedTopics(
	evt abi.Event,
	values []ExpectedWorkflowNameUpdatedTopics,
) ([]*evm.TopicValues, error) {
	var previousNameRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.PreviousName).IsZero() {
			previousNameRule = append(previousNameRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.PreviousName)
		if err != nil {
			return nil, err
		}
		previousNameRule = append(previousNameRule, fieldVal)
	}
	var newNameRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.NewName).IsZero() {
			newNameRule = append(newNameRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[1], v.NewName)
		if err != nil {
			return nil, err
		}
		newNameRule = append(newNameRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		previousNameRule,
		newNameRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeExpectedWorkflowNameUpdated decodes a log into a ExpectedWorkflowNameUpdated struct.
func (c *Codec) DecodeExpectedWorkflowNameUpdated(log *evm.Log) (*ExpectedWorkflowNameUpdatedDecoded, error) {
	event := new(ExpectedWorkflowNameUpdatedDecoded)
	if err := c.abi.UnpackIntoInterface(event, "ExpectedWorkflowNameUpdated", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["ExpectedWorkflowNameUpdated"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) ForwarderAddressUpdatedLogHash() []byte {
	return c.abi.Events["ForwarderAddressUpdated"].ID.Bytes()
}

func (c *Codec) EncodeForwarderAddressUpdatedTopics(
	evt abi.Event,
	values []ForwarderAddressUpdatedTopics,
) ([]*evm.TopicValues, error) {
	var previousForwarderRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.PreviousForwarder).IsZero() {
			previousForwarderRule = append(previousForwarderRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.PreviousForwarder)
		if err != nil {
			return nil, err
		}
		previousForwarderRule = append(previousForwarderRule, fieldVal)
	}
	var newForwarderRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.NewForwarder).IsZero() {
			newForwarderRule = append(newForwarderRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[1], v.NewForwarder)
		if err != nil {
			return nil, err
		}
		newForwarderRule = append(newForwarderRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		previousForwarderRule,
		newForwarderRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeForwarderAddressUpdated decodes a log into a ForwarderAddressUpdated struct.
func (c *Codec) DecodeForwarderAddressUpdated(log *evm.Log) (*ForwarderAddressUpdatedDecoded, error) {
	event := new(ForwarderAddressUpdatedDecoded)
	if err := c.abi.UnpackIntoInterface(event, "ForwarderAddressUpdated", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["ForwarderAddressUpdated"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) OwnershipHandoverCanceledLogHash() []byte {
	return c.abi.Events["OwnershipHandoverCanceled"].ID.Bytes()
}

func (c *Codec) EncodeOwnershipHandoverCanceledTopics(
	evt abi.Event,
	values []OwnershipHandoverCanceledTopics,
) ([]*evm.TopicValues, error) {
	var pendingOwnerRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.PendingOwner).IsZero() {
			pendingOwnerRule = append(pendingOwnerRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.PendingOwner)
		if err != nil {
			return nil, err
		}
		pendingOwnerRule = append(pendingOwnerRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		pendingOwnerRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeOwnershipHandoverCanceled decodes a log into a OwnershipHandoverCanceled struct.
func (c *Codec) DecodeOwnershipHandoverCanceled(log *evm.Log) (*OwnershipHandoverCanceledDecoded, error) {
	event := new(OwnershipHandoverCanceledDecoded)
	if err := c.abi.UnpackIntoInterface(event, "OwnershipHandoverCanceled", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["OwnershipHandoverCanceled"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) OwnershipHandoverRequestedLogHash() []byte {
	return c.abi.Events["OwnershipHandoverRequested"].ID.Bytes()
}

func (c *Codec) EncodeOwnershipHandoverRequestedTopics(
	evt abi.Event,
	values []OwnershipHandoverRequestedTopics,
) ([]*evm.TopicValues, error) {
	var pendingOwnerRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.PendingOwner).IsZero() {
			pendingOwnerRule = append(pendingOwnerRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.PendingOwner)
		if err != nil {
			return nil, err
		}
		pendingOwnerRule = append(pendingOwnerRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		pendingOwnerRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeOwnershipHandoverRequested decodes a log into a OwnershipHandoverRequested struct.
func (c *Codec) DecodeOwnershipHandoverRequested(log *evm.Log) (*OwnershipHandoverRequestedDecoded, error) {
	event := new(OwnershipHandoverRequestedDecoded)
	if err := c.abi.UnpackIntoInterface(event, "OwnershipHandoverRequested", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["OwnershipHandoverRequested"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) OwnershipTransferredLogHash() []byte {
	return c.abi.Events["OwnershipTransferred"].ID.Bytes()
}

func (c *Codec) EncodeOwnershipTransferredTopics(
	evt abi.Event,
	values []OwnershipTransferredTopics,
) ([]*evm.TopicValues, error) {
	var oldOwnerRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.OldOwner).IsZero() {
			oldOwnerRule = append(oldOwnerRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.OldOwner)
		if err != nil {
			return nil, err
		}
		oldOwnerRule = append(oldOwnerRule, fieldVal)
	}
	var newOwnerRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.NewOwner).IsZero() {
			newOwnerRule = append(newOwnerRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[1], v.NewOwner)
		if err != nil {
			return nil, err
		}
		newOwnerRule = append(newOwnerRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		oldOwnerRule,
		newOwnerRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeOwnershipTransferred decodes a log into a OwnershipTransferred struct.
func (c *Codec) DecodeOwnershipTransferred(log *evm.Log) (*OwnershipTransferredDecoded, error) {
	event := new(OwnershipTransferredDecoded)
	if err := c.abi.UnpackIntoInterface(event, "OwnershipTransferred", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["OwnershipTransferred"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) RelayerNodeSetLogHash() []byte {
	return c.abi.Events["RelayerNodeSet"].ID.Bytes()
}

func (c *Codec) EncodeRelayerNodeSetTopics(
	evt abi.Event,
	values []RelayerNodeSetTopics,
) ([]*evm.TopicValues, error) {
	var newRelayerNodeRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.NewRelayerNode).IsZero() {
			newRelayerNodeRule = append(newRelayerNodeRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[0], v.NewRelayerNode)
		if err != nil {
			return nil, err
		}
		newRelayerNodeRule = append(newRelayerNodeRule, fieldVal)
	}
	var oldRelayerNodeRule []interface{}
	for _, v := range values {
		if reflect.ValueOf(v.OldRelayerNode).IsZero() {
			oldRelayerNodeRule = append(oldRelayerNodeRule, common.Hash{})
			continue
		}
		fieldVal, err := bindings.PrepareTopicArg(evt.Inputs[1], v.OldRelayerNode)
		if err != nil {
			return nil, err
		}
		oldRelayerNodeRule = append(oldRelayerNodeRule, fieldVal)
	}

	rawTopics, err := abi.MakeTopics(
		newRelayerNodeRule,
		oldRelayerNodeRule,
	)
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeRelayerNodeSet decodes a log into a RelayerNodeSet struct.
func (c *Codec) DecodeRelayerNodeSet(log *evm.Log) (*RelayerNodeSetDecoded, error) {
	event := new(RelayerNodeSetDecoded)
	if err := c.abi.UnpackIntoInterface(event, "RelayerNodeSet", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["RelayerNodeSet"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c *Codec) SecurityWarningLogHash() []byte {
	return c.abi.Events["SecurityWarning"].ID.Bytes()
}

func (c *Codec) EncodeSecurityWarningTopics(
	evt abi.Event,
	values []SecurityWarningTopics,
) ([]*evm.TopicValues, error) {

	rawTopics, err := abi.MakeTopics()
	if err != nil {
		return nil, err
	}

	return bindings.PrepareTopics(rawTopics, evt.ID.Bytes()), nil
}

// DecodeSecurityWarning decodes a log into a SecurityWarning struct.
func (c *Codec) DecodeSecurityWarning(log *evm.Log) (*SecurityWarningDecoded, error) {
	event := new(SecurityWarningDecoded)
	if err := c.abi.UnpackIntoInterface(event, "SecurityWarning", log.Data); err != nil {
		return nil, err
	}
	var indexed abi.Arguments
	for _, arg := range c.abi.Events["SecurityWarning"].Inputs {
		if arg.Indexed {
			if arg.Type.T == abi.TupleTy {
				// abigen throws on tuple, so converting to bytes to
				// receive back the common.Hash as is instead of error
				arg.Type.T = abi.BytesTy
			}
			indexed = append(indexed, arg)
		}
	}
	// Convert [][]byte → []common.Hash
	topics := make([]common.Hash, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = common.BytesToHash(t)
	}

	if err := abi.ParseTopics(event, indexed, topics[1:]); err != nil {
		return nil, err
	}
	return event, nil
}

func (c IOracle) IfaPriceFeed(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[common.Address] {
	calldata, err := c.Codec.EncodeIfaPriceFeedMethodCall()
	if err != nil {
		return cre.PromiseFromResult[common.Address](*new(common.Address), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (common.Address, error) {
		return c.Codec.DecodeIfaPriceFeedMethodOutput(response.Data)
	})

}

func (c IOracle) GetExpectedAuthor(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[common.Address] {
	calldata, err := c.Codec.EncodeGetExpectedAuthorMethodCall()
	if err != nil {
		return cre.PromiseFromResult[common.Address](*new(common.Address), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (common.Address, error) {
		return c.Codec.DecodeGetExpectedAuthorMethodOutput(response.Data)
	})

}

func (c IOracle) GetExpectedWorkflowId(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[[32]byte] {
	calldata, err := c.Codec.EncodeGetExpectedWorkflowIdMethodCall()
	if err != nil {
		return cre.PromiseFromResult[[32]byte](*new([32]byte), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) ([32]byte, error) {
		return c.Codec.DecodeGetExpectedWorkflowIdMethodOutput(response.Data)
	})

}

func (c IOracle) GetExpectedWorkflowName(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[[10]byte] {
	calldata, err := c.Codec.EncodeGetExpectedWorkflowNameMethodCall()
	if err != nil {
		return cre.PromiseFromResult[[10]byte](*new([10]byte), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) ([10]byte, error) {
		return c.Codec.DecodeGetExpectedWorkflowNameMethodOutput(response.Data)
	})

}

func (c IOracle) GetForwarderAddress(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[common.Address] {
	calldata, err := c.Codec.EncodeGetForwarderAddressMethodCall()
	if err != nil {
		return cre.PromiseFromResult[common.Address](*new(common.Address), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (common.Address, error) {
		return c.Codec.DecodeGetForwarderAddressMethodOutput(response.Data)
	})

}

func (c IOracle) Owner(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[common.Address] {
	calldata, err := c.Codec.EncodeOwnerMethodCall()
	if err != nil {
		return cre.PromiseFromResult[common.Address](*new(common.Address), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (common.Address, error) {
		return c.Codec.DecodeOwnerMethodOutput(response.Data)
	})

}

func (c IOracle) OwnershipHandoverExpiresAt(
	runtime cre.Runtime,
	args OwnershipHandoverExpiresAtInput,
	blockNumber *big.Int,
) cre.Promise[*big.Int] {
	calldata, err := c.Codec.EncodeOwnershipHandoverExpiresAtMethodCall(args)
	if err != nil {
		return cre.PromiseFromResult[*big.Int](*new(*big.Int), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (*big.Int, error) {
		return c.Codec.DecodeOwnershipHandoverExpiresAtMethodOutput(response.Data)
	})

}

func (c IOracle) RelayerNode(
	runtime cre.Runtime,
	blockNumber *big.Int,
) cre.Promise[common.Address] {
	calldata, err := c.Codec.EncodeRelayerNodeMethodCall()
	if err != nil {
		return cre.PromiseFromResult[common.Address](*new(common.Address), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (common.Address, error) {
		return c.Codec.DecodeRelayerNodeMethodOutput(response.Data)
	})

}

func (c IOracle) SupportsInterface(
	runtime cre.Runtime,
	args SupportsInterfaceInput,
	blockNumber *big.Int,
) cre.Promise[bool] {
	calldata, err := c.Codec.EncodeSupportsInterfaceMethodCall(args)
	if err != nil {
		return cre.PromiseFromResult[bool](*new(bool), err)
	}

	var bn cre.Promise[*pb.BigInt]
	if blockNumber == nil {
		promise := c.client.HeaderByNumber(runtime, &evm.HeaderByNumberRequest{
			BlockNumber: bindings.FinalizedBlockNumber,
		})

		bn = cre.Then(promise, func(finalizedBlock *evm.HeaderByNumberReply) (*pb.BigInt, error) {
			if finalizedBlock == nil || finalizedBlock.Header == nil {
				return nil, errors.New("failed to get finalized block header")
			}
			return finalizedBlock.Header.BlockNumber, nil
		})
	} else {
		bn = cre.PromiseFromResult(pb.NewBigIntFromInt(blockNumber), nil)
	}

	promise := cre.ThenPromise(bn, func(bn *pb.BigInt) cre.Promise[*evm.CallContractReply] {
		return c.client.CallContract(runtime, &evm.CallContractRequest{
			Call:        &evm.CallMsg{To: c.Address.Bytes(), Data: calldata},
			BlockNumber: bn,
		})
	})
	return cre.Then(promise, func(response *evm.CallContractReply) (bool, error) {
		return c.Codec.DecodeSupportsInterfaceMethodOutput(response.Data)
	})

}

func (c IOracle) WriteReportFromIIfaPriceFeedPriceFeed(
	runtime cre.Runtime,
	input IIfaPriceFeedPriceFeed,
	gasConfig *evm.GasConfig,
) cre.Promise[*evm.WriteReportReply] {
	encoded, err := c.Codec.EncodeIIfaPriceFeedPriceFeedStruct(input)
	if err != nil {
		return cre.PromiseFromResult[*evm.WriteReportReply](nil, err)
	}
	promise := runtime.GenerateReport(&pb2.ReportRequest{
		EncodedPayload: encoded,
		EncoderName:    "evm",
		SigningAlgo:    "ecdsa",
		HashingAlgo:    "keccak256",
	})

	return cre.ThenPromise(promise, func(report *cre.Report) cre.Promise[*evm.WriteReportReply] {
		return c.client.WriteReport(runtime, &evm.WriteCreReportRequest{
			Receiver:  c.Address.Bytes(),
			Report:    report,
			GasConfig: gasConfig,
		})
	})
}

func (c IOracle) WriteReportFromIfaPriceFeedVerifierSumissionData(
	runtime cre.Runtime,
	input IfaPriceFeedVerifierSumissionData,
	gasConfig *evm.GasConfig,
) cre.Promise[*evm.WriteReportReply] {
	encoded, err := c.Codec.EncodeIfaPriceFeedVerifierSumissionDataStruct(input)
	if err != nil {
		return cre.PromiseFromResult[*evm.WriteReportReply](nil, err)
	}
	promise := runtime.GenerateReport(&pb2.ReportRequest{
		EncodedPayload: encoded,
		EncoderName:    "evm",
		SigningAlgo:    "ecdsa",
		HashingAlgo:    "keccak256",
	})

	return cre.ThenPromise(promise, func(report *cre.Report) cre.Promise[*evm.WriteReportReply] {
		return c.client.WriteReport(runtime, &evm.WriteCreReportRequest{
			Receiver:  c.Address.Bytes(),
			Report:    report,
			GasConfig: gasConfig,
		})
	})
}

func (c IOracle) WriteReport(
	runtime cre.Runtime,
	report *cre.Report,
	gasConfig *evm.GasConfig,
) cre.Promise[*evm.WriteReportReply] {
	return c.client.WriteReport(runtime, &evm.WriteCreReportRequest{
		Receiver:  c.Address.Bytes(),
		Report:    report,
		GasConfig: gasConfig,
	})
}

// DecodeAlreadyInitializedError decodes a AlreadyInitialized error from revert data.
func (c *IOracle) DecodeAlreadyInitializedError(data []byte) (*AlreadyInitialized, error) {
	args := c.ABI.Errors["AlreadyInitialized"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &AlreadyInitialized{}, nil
}

// Error implements the error interface for AlreadyInitialized.
func (e *AlreadyInitialized) Error() string {
	return fmt.Sprintf("AlreadyInitialized error:")
}

// DecodeInvalidAssePriceError decodes a InvalidAssePrice error from revert data.
func (c *IOracle) DecodeInvalidAssePriceError(data []byte) (*InvalidAssePrice, error) {
	args := c.ABI.Errors["InvalidAssePrice"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &InvalidAssePrice{}, nil
}

// Error implements the error interface for InvalidAssePrice.
func (e *InvalidAssePrice) Error() string {
	return fmt.Sprintf("InvalidAssePrice error:")
}

// DecodeInvalidAssetIndexorPriceLengthError decodes a InvalidAssetIndexorPriceLength error from revert data.
func (c *IOracle) DecodeInvalidAssetIndexorPriceLengthError(data []byte) (*InvalidAssetIndexorPriceLength, error) {
	args := c.ABI.Errors["InvalidAssetIndexorPriceLength"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &InvalidAssetIndexorPriceLength{}, nil
}

// Error implements the error interface for InvalidAssetIndexorPriceLength.
func (e *InvalidAssetIndexorPriceLength) Error() string {
	return fmt.Sprintf("InvalidAssetIndexorPriceLength error:")
}

// DecodeInvalidAuthorError decodes a InvalidAuthor error from revert data.
func (c *IOracle) DecodeInvalidAuthorError(data []byte) (*InvalidAuthor, error) {
	args := c.ABI.Errors["InvalidAuthor"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values, got %d", len(values))
	}

	received, ok0 := values[0].(common.Address)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for received in InvalidAuthor error")
	}

	expected, ok1 := values[1].(common.Address)
	if !ok1 {
		return nil, fmt.Errorf("unexpected type for expected in InvalidAuthor error")
	}

	return &InvalidAuthor{
		Received: received,
		Expected: expected,
	}, nil
}

// Error implements the error interface for InvalidAuthor.
func (e *InvalidAuthor) Error() string {
	return fmt.Sprintf("InvalidAuthor error: received=%v; expected=%v;", e.Received, e.Expected)
}

// DecodeInvalidForwarderAddressError decodes a InvalidForwarderAddress error from revert data.
func (c *IOracle) DecodeInvalidForwarderAddressError(data []byte) (*InvalidForwarderAddress, error) {
	args := c.ABI.Errors["InvalidForwarderAddress"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &InvalidForwarderAddress{}, nil
}

// Error implements the error interface for InvalidForwarderAddress.
func (e *InvalidForwarderAddress) Error() string {
	return fmt.Sprintf("InvalidForwarderAddress error:")
}

// DecodeInvalidRelayerNodeError decodes a InvalidRelayerNode error from revert data.
func (c *IOracle) DecodeInvalidRelayerNodeError(data []byte) (*InvalidRelayerNode, error) {
	args := c.ABI.Errors["InvalidRelayerNode"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("expected 1 values, got %d", len(values))
	}

	address, ok0 := values[0].(common.Address)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for address in InvalidRelayerNode error")
	}

	return &InvalidRelayerNode{
		Address: address,
	}, nil
}

// Error implements the error interface for InvalidRelayerNode.
func (e *InvalidRelayerNode) Error() string {
	return fmt.Sprintf("InvalidRelayerNode error: address=%v;", e.Address)
}

// DecodeInvalidSenderError decodes a InvalidSender error from revert data.
func (c *IOracle) DecodeInvalidSenderError(data []byte) (*InvalidSender, error) {
	args := c.ABI.Errors["InvalidSender"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values, got %d", len(values))
	}

	sender, ok0 := values[0].(common.Address)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for sender in InvalidSender error")
	}

	expected, ok1 := values[1].(common.Address)
	if !ok1 {
		return nil, fmt.Errorf("unexpected type for expected in InvalidSender error")
	}

	return &InvalidSender{
		Sender:   sender,
		Expected: expected,
	}, nil
}

// Error implements the error interface for InvalidSender.
func (e *InvalidSender) Error() string {
	return fmt.Sprintf("InvalidSender error: sender=%v; expected=%v;", e.Sender, e.Expected)
}

// DecodeInvalidWorkflowIdError decodes a InvalidWorkflowId error from revert data.
func (c *IOracle) DecodeInvalidWorkflowIdError(data []byte) (*InvalidWorkflowId, error) {
	args := c.ABI.Errors["InvalidWorkflowId"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values, got %d", len(values))
	}

	received, ok0 := values[0].([32]byte)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for received in InvalidWorkflowId error")
	}

	expected, ok1 := values[1].([32]byte)
	if !ok1 {
		return nil, fmt.Errorf("unexpected type for expected in InvalidWorkflowId error")
	}

	return &InvalidWorkflowId{
		Received: received,
		Expected: expected,
	}, nil
}

// Error implements the error interface for InvalidWorkflowId.
func (e *InvalidWorkflowId) Error() string {
	return fmt.Sprintf("InvalidWorkflowId error: received=%v; expected=%v;", e.Received, e.Expected)
}

// DecodeInvalidWorkflowNameError decodes a InvalidWorkflowName error from revert data.
func (c *IOracle) DecodeInvalidWorkflowNameError(data []byte) (*InvalidWorkflowName, error) {
	args := c.ABI.Errors["InvalidWorkflowName"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values, got %d", len(values))
	}

	received, ok0 := values[0].([10]byte)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for received in InvalidWorkflowName error")
	}

	expected, ok1 := values[1].([10]byte)
	if !ok1 {
		return nil, fmt.Errorf("unexpected type for expected in InvalidWorkflowName error")
	}

	return &InvalidWorkflowName{
		Received: received,
		Expected: expected,
	}, nil
}

// Error implements the error interface for InvalidWorkflowName.
func (e *InvalidWorkflowName) Error() string {
	return fmt.Sprintf("InvalidWorkflowName error: received=%v; expected=%v;", e.Received, e.Expected)
}

// DecodeNewOwnerIsZeroAddressError decodes a NewOwnerIsZeroAddress error from revert data.
func (c *IOracle) DecodeNewOwnerIsZeroAddressError(data []byte) (*NewOwnerIsZeroAddress, error) {
	args := c.ABI.Errors["NewOwnerIsZeroAddress"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &NewOwnerIsZeroAddress{}, nil
}

// Error implements the error interface for NewOwnerIsZeroAddress.
func (e *NewOwnerIsZeroAddress) Error() string {
	return fmt.Sprintf("NewOwnerIsZeroAddress error:")
}

// DecodeNoHandoverRequestError decodes a NoHandoverRequest error from revert data.
func (c *IOracle) DecodeNoHandoverRequestError(data []byte) (*NoHandoverRequest, error) {
	args := c.ABI.Errors["NoHandoverRequest"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &NoHandoverRequest{}, nil
}

// Error implements the error interface for NoHandoverRequest.
func (e *NoHandoverRequest) Error() string {
	return fmt.Sprintf("NoHandoverRequest error:")
}

// DecodeOnlyRelayerNodeError decodes a OnlyRelayerNode error from revert data.
func (c *IOracle) DecodeOnlyRelayerNodeError(data []byte) (*OnlyRelayerNode, error) {
	args := c.ABI.Errors["OnlyRelayerNode"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("expected 1 values, got %d", len(values))
	}

	caller, ok0 := values[0].(common.Address)
	if !ok0 {
		return nil, fmt.Errorf("unexpected type for caller in OnlyRelayerNode error")
	}

	return &OnlyRelayerNode{
		Caller: caller,
	}, nil
}

// Error implements the error interface for OnlyRelayerNode.
func (e *OnlyRelayerNode) Error() string {
	return fmt.Sprintf("OnlyRelayerNode error: caller=%v;", e.Caller)
}

// DecodeUnauthorizedError decodes a Unauthorized error from revert data.
func (c *IOracle) DecodeUnauthorizedError(data []byte) (*Unauthorized, error) {
	args := c.ABI.Errors["Unauthorized"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &Unauthorized{}, nil
}

// Error implements the error interface for Unauthorized.
func (e *Unauthorized) Error() string {
	return fmt.Sprintf("Unauthorized error:")
}

// DecodeWorkflowNameRequiresAuthorValidationError decodes a WorkflowNameRequiresAuthorValidation error from revert data.
func (c *IOracle) DecodeWorkflowNameRequiresAuthorValidationError(data []byte) (*WorkflowNameRequiresAuthorValidation, error) {
	args := c.ABI.Errors["WorkflowNameRequiresAuthorValidation"].Inputs
	values, err := args.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("failed to unpack error: %w", err)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("expected 0 values, got %d", len(values))
	}

	return &WorkflowNameRequiresAuthorValidation{}, nil
}

// Error implements the error interface for WorkflowNameRequiresAuthorValidation.
func (e *WorkflowNameRequiresAuthorValidation) Error() string {
	return fmt.Sprintf("WorkflowNameRequiresAuthorValidation error:")
}

func (c *IOracle) UnpackError(data []byte) (any, error) {
	switch common.Bytes2Hex(data[:4]) {
	case common.Bytes2Hex(c.ABI.Errors["AlreadyInitialized"].ID.Bytes()[:4]):
		return c.DecodeAlreadyInitializedError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidAssePrice"].ID.Bytes()[:4]):
		return c.DecodeInvalidAssePriceError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidAssetIndexorPriceLength"].ID.Bytes()[:4]):
		return c.DecodeInvalidAssetIndexorPriceLengthError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidAuthor"].ID.Bytes()[:4]):
		return c.DecodeInvalidAuthorError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidForwarderAddress"].ID.Bytes()[:4]):
		return c.DecodeInvalidForwarderAddressError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidRelayerNode"].ID.Bytes()[:4]):
		return c.DecodeInvalidRelayerNodeError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidSender"].ID.Bytes()[:4]):
		return c.DecodeInvalidSenderError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidWorkflowId"].ID.Bytes()[:4]):
		return c.DecodeInvalidWorkflowIdError(data)
	case common.Bytes2Hex(c.ABI.Errors["InvalidWorkflowName"].ID.Bytes()[:4]):
		return c.DecodeInvalidWorkflowNameError(data)
	case common.Bytes2Hex(c.ABI.Errors["NewOwnerIsZeroAddress"].ID.Bytes()[:4]):
		return c.DecodeNewOwnerIsZeroAddressError(data)
	case common.Bytes2Hex(c.ABI.Errors["NoHandoverRequest"].ID.Bytes()[:4]):
		return c.DecodeNoHandoverRequestError(data)
	case common.Bytes2Hex(c.ABI.Errors["OnlyRelayerNode"].ID.Bytes()[:4]):
		return c.DecodeOnlyRelayerNodeError(data)
	case common.Bytes2Hex(c.ABI.Errors["Unauthorized"].ID.Bytes()[:4]):
		return c.DecodeUnauthorizedError(data)
	case common.Bytes2Hex(c.ABI.Errors["WorkflowNameRequiresAuthorValidation"].ID.Bytes()[:4]):
		return c.DecodeWorkflowNameRequiresAuthorValidationError(data)
	default:
		return nil, errors.New("unknown error selector")
	}
}

// ExpectedAuthorUpdatedTrigger wraps the raw log trigger and provides decoded ExpectedAuthorUpdatedDecoded data
type ExpectedAuthorUpdatedTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into ExpectedAuthorUpdated data
func (t *ExpectedAuthorUpdatedTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[ExpectedAuthorUpdatedDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeExpectedAuthorUpdated(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ExpectedAuthorUpdated log: %w", err)
	}

	return &bindings.DecodedLog[ExpectedAuthorUpdatedDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerExpectedAuthorUpdatedLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []ExpectedAuthorUpdatedTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[ExpectedAuthorUpdatedDecoded]], error) {
	event := c.ABI.Events["ExpectedAuthorUpdated"]
	topics, err := c.Codec.EncodeExpectedAuthorUpdatedTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for ExpectedAuthorUpdated: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &ExpectedAuthorUpdatedTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsExpectedAuthorUpdated(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.ExpectedAuthorUpdatedLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// ExpectedWorkflowIdUpdatedTrigger wraps the raw log trigger and provides decoded ExpectedWorkflowIdUpdatedDecoded data
type ExpectedWorkflowIdUpdatedTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into ExpectedWorkflowIdUpdated data
func (t *ExpectedWorkflowIdUpdatedTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[ExpectedWorkflowIdUpdatedDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeExpectedWorkflowIdUpdated(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ExpectedWorkflowIdUpdated log: %w", err)
	}

	return &bindings.DecodedLog[ExpectedWorkflowIdUpdatedDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerExpectedWorkflowIdUpdatedLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []ExpectedWorkflowIdUpdatedTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[ExpectedWorkflowIdUpdatedDecoded]], error) {
	event := c.ABI.Events["ExpectedWorkflowIdUpdated"]
	topics, err := c.Codec.EncodeExpectedWorkflowIdUpdatedTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for ExpectedWorkflowIdUpdated: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &ExpectedWorkflowIdUpdatedTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsExpectedWorkflowIdUpdated(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.ExpectedWorkflowIdUpdatedLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// ExpectedWorkflowNameUpdatedTrigger wraps the raw log trigger and provides decoded ExpectedWorkflowNameUpdatedDecoded data
type ExpectedWorkflowNameUpdatedTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into ExpectedWorkflowNameUpdated data
func (t *ExpectedWorkflowNameUpdatedTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[ExpectedWorkflowNameUpdatedDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeExpectedWorkflowNameUpdated(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ExpectedWorkflowNameUpdated log: %w", err)
	}

	return &bindings.DecodedLog[ExpectedWorkflowNameUpdatedDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerExpectedWorkflowNameUpdatedLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []ExpectedWorkflowNameUpdatedTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[ExpectedWorkflowNameUpdatedDecoded]], error) {
	event := c.ABI.Events["ExpectedWorkflowNameUpdated"]
	topics, err := c.Codec.EncodeExpectedWorkflowNameUpdatedTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for ExpectedWorkflowNameUpdated: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &ExpectedWorkflowNameUpdatedTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsExpectedWorkflowNameUpdated(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.ExpectedWorkflowNameUpdatedLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// ForwarderAddressUpdatedTrigger wraps the raw log trigger and provides decoded ForwarderAddressUpdatedDecoded data
type ForwarderAddressUpdatedTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into ForwarderAddressUpdated data
func (t *ForwarderAddressUpdatedTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[ForwarderAddressUpdatedDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeForwarderAddressUpdated(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ForwarderAddressUpdated log: %w", err)
	}

	return &bindings.DecodedLog[ForwarderAddressUpdatedDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerForwarderAddressUpdatedLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []ForwarderAddressUpdatedTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[ForwarderAddressUpdatedDecoded]], error) {
	event := c.ABI.Events["ForwarderAddressUpdated"]
	topics, err := c.Codec.EncodeForwarderAddressUpdatedTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for ForwarderAddressUpdated: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &ForwarderAddressUpdatedTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsForwarderAddressUpdated(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.ForwarderAddressUpdatedLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// OwnershipHandoverCanceledTrigger wraps the raw log trigger and provides decoded OwnershipHandoverCanceledDecoded data
type OwnershipHandoverCanceledTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into OwnershipHandoverCanceled data
func (t *OwnershipHandoverCanceledTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[OwnershipHandoverCanceledDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeOwnershipHandoverCanceled(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode OwnershipHandoverCanceled log: %w", err)
	}

	return &bindings.DecodedLog[OwnershipHandoverCanceledDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerOwnershipHandoverCanceledLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []OwnershipHandoverCanceledTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[OwnershipHandoverCanceledDecoded]], error) {
	event := c.ABI.Events["OwnershipHandoverCanceled"]
	topics, err := c.Codec.EncodeOwnershipHandoverCanceledTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for OwnershipHandoverCanceled: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &OwnershipHandoverCanceledTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsOwnershipHandoverCanceled(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.OwnershipHandoverCanceledLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// OwnershipHandoverRequestedTrigger wraps the raw log trigger and provides decoded OwnershipHandoverRequestedDecoded data
type OwnershipHandoverRequestedTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into OwnershipHandoverRequested data
func (t *OwnershipHandoverRequestedTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[OwnershipHandoverRequestedDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeOwnershipHandoverRequested(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode OwnershipHandoverRequested log: %w", err)
	}

	return &bindings.DecodedLog[OwnershipHandoverRequestedDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerOwnershipHandoverRequestedLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []OwnershipHandoverRequestedTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[OwnershipHandoverRequestedDecoded]], error) {
	event := c.ABI.Events["OwnershipHandoverRequested"]
	topics, err := c.Codec.EncodeOwnershipHandoverRequestedTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for OwnershipHandoverRequested: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &OwnershipHandoverRequestedTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsOwnershipHandoverRequested(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.OwnershipHandoverRequestedLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// OwnershipTransferredTrigger wraps the raw log trigger and provides decoded OwnershipTransferredDecoded data
type OwnershipTransferredTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into OwnershipTransferred data
func (t *OwnershipTransferredTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[OwnershipTransferredDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeOwnershipTransferred(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode OwnershipTransferred log: %w", err)
	}

	return &bindings.DecodedLog[OwnershipTransferredDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerOwnershipTransferredLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []OwnershipTransferredTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[OwnershipTransferredDecoded]], error) {
	event := c.ABI.Events["OwnershipTransferred"]
	topics, err := c.Codec.EncodeOwnershipTransferredTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for OwnershipTransferred: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &OwnershipTransferredTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsOwnershipTransferred(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.OwnershipTransferredLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// RelayerNodeSetTrigger wraps the raw log trigger and provides decoded RelayerNodeSetDecoded data
type RelayerNodeSetTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into RelayerNodeSet data
func (t *RelayerNodeSetTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[RelayerNodeSetDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeRelayerNodeSet(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RelayerNodeSet log: %w", err)
	}

	return &bindings.DecodedLog[RelayerNodeSetDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerRelayerNodeSetLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []RelayerNodeSetTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[RelayerNodeSetDecoded]], error) {
	event := c.ABI.Events["RelayerNodeSet"]
	topics, err := c.Codec.EncodeRelayerNodeSetTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for RelayerNodeSet: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &RelayerNodeSetTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsRelayerNodeSet(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.RelayerNodeSetLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}

// SecurityWarningTrigger wraps the raw log trigger and provides decoded SecurityWarningDecoded data
type SecurityWarningTrigger struct {
	cre.Trigger[*evm.Log, *evm.Log]          // Embed the raw trigger
	contract                        *IOracle // Keep reference for decoding
}

// Adapt method that decodes the log into SecurityWarning data
func (t *SecurityWarningTrigger) Adapt(l *evm.Log) (*bindings.DecodedLog[SecurityWarningDecoded], error) {
	// Decode the log using the contract's codec
	decoded, err := t.contract.Codec.DecodeSecurityWarning(l)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SecurityWarning log: %w", err)
	}

	return &bindings.DecodedLog[SecurityWarningDecoded]{
		Log:  l,        // Original log
		Data: *decoded, // Decoded data
	}, nil
}

func (c *IOracle) LogTriggerSecurityWarningLog(chainSelector uint64, confidence evm.ConfidenceLevel, filters []SecurityWarningTopics) (cre.Trigger[*evm.Log, *bindings.DecodedLog[SecurityWarningDecoded]], error) {
	event := c.ABI.Events["SecurityWarning"]
	topics, err := c.Codec.EncodeSecurityWarningTopics(event, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to encode topics for SecurityWarning: %w", err)
	}

	rawTrigger := evm.LogTrigger(chainSelector, &evm.FilterLogTriggerRequest{
		Addresses:  [][]byte{c.Address.Bytes()},
		Topics:     topics,
		Confidence: confidence,
	})

	return &SecurityWarningTrigger{
		Trigger:  rawTrigger,
		contract: c,
	}, nil
}

func (c *IOracle) FilterLogsSecurityWarning(runtime cre.Runtime, options *bindings.FilterOptions) (cre.Promise[*evm.FilterLogsReply], error) {
	if options == nil {
		return nil, errors.New("FilterLogs options are required.")
	}
	return c.client.FilterLogs(runtime, &evm.FilterLogsRequest{
		FilterQuery: &evm.FilterQuery{
			Addresses: [][]byte{c.Address.Bytes()},
			Topics: []*evm.Topics{
				{Topic: [][]byte{c.Codec.SecurityWarningLogHash()}},
			},
			BlockHash: options.BlockHash,
			FromBlock: pb.NewBigIntFromInt(options.FromBlock),
			ToBlock:   pb.NewBigIntFromInt(options.ToBlock),
		},
	}), nil
}
