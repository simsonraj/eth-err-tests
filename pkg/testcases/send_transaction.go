package testcases

import (
	"context"
	"fmt"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SendTransactionTestCase struct{}

func (t *SendTransactionTestCase) Name() string {
	return "eth_sendRawTransaction"
}

func (t *SendTransactionTestCase) RequiresContract() bool {
	return true
}

func (t *SendTransactionTestCase) GetRequests(cfg config.Config) []pkgTypes.Meta {
	scenarios := GetScenarios(cfg)
	requests := make([]pkgTypes.Meta, 0, len(scenarios))

	for _, scenario := range scenarios {
		requests = append(requests, pkgTypes.Meta{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      scenario.ID,
				Method:  scenario.Method,
				Params:  []interface{}{},
			},
			Desc: scenario.Desc,
		})
	}

	return requests
}

func (t *SendTransactionTestCase) Execute(cfg config.Config) {
	ctx := context.Background()

	// Connect to the Ethereum client
	client, err := ethclient.Dial(cfg.Url)
	if err != nil {
		fmt.Println("Error connecting to Ethereum client:", err)
		return
	}
	defer client.Close()

	// Get all scenarios with config
	scenarios := GetScenarios(cfg)

	// Execute each scenario
	for _, scenario := range scenarios {
		if err := jsonrpc.SendTransaction(ctx, client, cfg, scenario); err != nil {
			fmt.Printf("Error executing scenario %d (%s): %v\n", scenario.ID, scenario.Desc, err)
		}
		fmt.Println()
	}
}

func (t *SendTransactionTestCase) corruptTransaction(signedTx *types.Transaction, params *pkgTypes.TxParams) *types.Transaction {
	// Corrupt the last byte of the data
	data := signedTx.Data()
	if len(data) > 0 {
		corruptIndex := len(data) - 1
		corruptedData := make([]byte, len(data))
		copy(corruptedData, data)
		corruptedData[corruptIndex] = corruptedData[corruptIndex] ^ 0xff
		return types.NewTransaction(signedTx.Nonce(), *params.To, signedTx.Value(), signedTx.Gas(), signedTx.GasPrice(), corruptedData)
	}
	return signedTx
}

func NewSendTransactionTestCase() pkgTypes.TestCase {
	return &SendTransactionTestCase{}
}
