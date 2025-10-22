package testcases

import (
	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
)

type CallTestCase struct{}

func (t *CallTestCase) Name() string {
	return "eth_call"
}

func (t *CallTestCase) RequiresContract() bool {
	return true
}

func (t *CallTestCase) GetRequests(cfg config.Config) []pkgTypes.Meta {
	return []pkgTypes.Meta{
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      1,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to":   cfg.ToContract,
						"data": "0x2e64cec1",
					},
					"latest",
				},
			},
			Desc: "Proper request",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      3,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to":   cfg.InvalidContract,
						"data": "0x2e64cec1",
					},
					"latest",
				}},
			Desc: "Invalid contract",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      4,
				Method:  "eth_call",
				Params:  []interface{}{"latest"}},
			Desc: "Invalid params types",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      5,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to": "0x1234",
					},
					"latest",
				}},
			Desc: "Invalid 1st argument",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      6,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					},
					"unsupported",
				}},
			Desc: "Invalid 2nd argument",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      7,
				Method:  "eth_wrongCall",
				Params: []interface{}{
					map[string]string{
						"to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					},
					"latest",
				}},
			Desc: "Incorrect method name",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      8,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					},
					"latest",
				}},
			Desc: "Missing 'data' field for contract function call",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      9,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"data": "0x123invalid",
					},
					"latest",
				}},
			Desc: "Invalid 'data' field",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      10,
				Method:  "eth_call",
				Params: []interface{}{
					map[string]string{
						"to":   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
						"data": "0x12345678",
					},
					0xabcdef,
				},
			},
			Desc: "Invalid block parameter format",
		},
	}
}

func (t *CallTestCase) Execute(cfg config.Config) {
	requests := t.GetRequests(cfg)
	jsonrpc.SendReq(requests, cfg)
}

func NewCallTestCase() pkgTypes.TestCase {
	return &CallTestCase{}
}
