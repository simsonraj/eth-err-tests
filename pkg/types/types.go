package types

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/eth-error-tests/pkg/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TestCase interface {
	Name() string
	RequiresContract() bool
	GetRequests(cfg config.Config) []Meta
	Execute(cfg config.Config)
}

// TestSuite represents a group of related test cases
type TestSuite struct {
	Name              string
	Description       string
	TestCases         []TestCase
	RequiresContracts bool
}

type DeploymentResult struct {
	ContractName    string
	ContractAddress common.Address
	TxHash          common.Hash
	Success         bool
	Error           error
}

type TestContext struct {
	Config            config.Config
	DeployedContracts map[string]common.Address // contract name -> address
	ClientName        string                    // e.g., "besu", "geth", "reth"
}

type TestResult struct {
	TestName string
	Scenario string
	Request  string
	Response string
	Success  bool
	Error    error
}

type Scenario struct {
	ID        int
	Desc      string
	Method    string
	Modifiers []Modifier
	PreSend   func(ctx context.Context, client *ethclient.Client, cfg config.Config, params *TxParams) (string, error) // Returns first raw tx for batch
	UseBatch  bool                                                                                                     // If true, PreSend should return a raw transaction to send in batch
}

type TxParams struct {
	Nonce       uint64
	To          *common.Address
	Value       *big.Int
	Data        []byte
	Gas         uint64
	GasPrice    *big.Int
	GasTipCap   *big.Int
	GasFeeCap   *big.Int
	IsDynamic   bool
	ChainID     int64
	PrivateKey  *ecdsa.PrivateKey
	FromAddress common.Address
}

// Modifier is a function that modifies transaction parameters to simulate different scenarios
type Modifier func(ctx context.Context, client *ethclient.Client, params *TxParams) error

type Meta struct {
	JsonRpcRequest `json:"jsonrpc"`
	Desc           string `json:"desc"`
}

type JsonRpcRequest struct {
	JsonRpc string        `json:"jsonrpc"`
	Id      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}
