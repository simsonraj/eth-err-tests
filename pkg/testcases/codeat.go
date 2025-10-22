package testcases

import (
	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
)

type CodeAtTestCase struct{}

func (t *CodeAtTestCase) Name() string {
	return "eth_getCode"
}

func (t *CodeAtTestCase) RequiresContract() bool {
	return true
}

func (t *CodeAtTestCase) GetRequests(cfg config.Config) []pkgTypes.Meta {
	return []pkgTypes.Meta{
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      1,
				Method:  "eth_getCode",
				Params:  []interface{}{cfg.ToContract, "latest"},
			},
			Desc: "Proper request",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      2,
				Method:  "eth_getCode",
				Params:  []interface{}{cfg.InvalidContract, "latest"},
			},
			Desc: "Invalid contract address",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      3,
				Method:  "eth_getCode",
				Params:  []interface{}{cfg.ToContract, "unsupported"},
			},
			Desc: "Unsupported block parameter",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      4,
				Method:  "eth_wrongGetCode",
				Params:  []interface{}{cfg.ToContract, "latest"},
			},
			Desc: "Incorrect method name",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      5,
				Method:  "eth_getCode",
				Params:  []interface{}{0xabcdef, "latest"},
			},
			Desc: "Invalid contract address format",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      6,
				Method:  "eth_getCode",
				Params:  []interface{}{cfg.ToContract, 0xabcdef},
			},
			Desc: "Invalid block parameter format",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      7,
				Method:  "eth_getCode",
				Params:  []interface{}{cfg.ToContract, "latest", "latest"},
			},
			Desc: "Too many arguments",
		},
	}
}

func (t *CodeAtTestCase) Execute(cfg config.Config) {
	requests := t.GetRequests(cfg)
	jsonrpc.SendReq(requests, cfg)
}

func NewCodeAtTestCase() pkgTypes.TestCase {
	return &CodeAtTestCase{}
}
