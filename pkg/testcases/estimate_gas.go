package testcases

import (
	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
)

type EstimateGasTestCase struct{}

func (t *EstimateGasTestCase) Name() string {
	return "eth_estimateGas"
}

func (t *EstimateGasTestCase) RequiresContract() bool {
	return true
}

func (t *EstimateGasTestCase) GetRequests(cfg config.Config) []pkgTypes.Meta {
	return []pkgTypes.Meta{
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      1,
				Method:  "eth_estimateGas",
				Params: []interface{}{
					map[string]string{
						"from": cfg.From,
						"to":   cfg.ToContract,
						"data": "0x6057361d0002c6c8f4b6852b7fe72b4cbf8d304a8b4f8b1eec216b77e1284cc4",
					},
				},
			},
			Desc: "Proper request",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      3,
				Method:  "eth_estimateGas",
				Params: []interface{}{
					map[string]string{
						"from": cfg.From,
						"to":   cfg.InvalidContract,
						"data": "0x6057361d0002c6c8f4b6852b7fe72b4cbf8d304a8b4f8b1eec216b77e1284cc4",
					},
				},
			},
			Desc: "Invalid contract",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      4,
				Method:  "eth_estimateGas",
				Params:  []interface{}{"latest"},
			},
			Desc: "Invalid params types",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      5,
				Method:  "eth_wrongEstimateGas",
				Params: []interface{}{
					map[string]string{
						"from": cfg.From,
						"to":   cfg.ToContract,
						"data": "0x6057361d0002c6c8f4b6852b7fe72b4cbf8d304a8b4f8b1eec216b77e1284cc4",
					},
				},
			},
			Desc: "Incorrect method name",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      6,
				Method:  "eth_estimateGas",
				Params: []interface{}{
					map[string]string{
						"to": "0x1234",
					},
				},
			},
			Desc: "Invalid 1st argument",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      7,
				Method:  "eth_estimateGas",
				Params: []interface{}{
					map[string]string{
						"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"data": "0x123invalid",
					},
				},
			},
			Desc: "Invalid data field",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      8,
				Method:  "eth_estimateGas",
				Params: []interface{}{
					map[string]string{
						"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"data": "0x12345678",
					},
				},
			},
			Desc: "Missing from field",
		},
	}
}

func (t *EstimateGasTestCase) Execute(cfg config.Config) {
	requests := t.GetRequests(cfg)
	jsonrpc.SendReq(requests, cfg)
}

func NewEstimateGasTestCase() pkgTypes.TestCase {
	return &EstimateGasTestCase{}
}
