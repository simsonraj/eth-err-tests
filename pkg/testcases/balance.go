package testcases

import (
	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
)

type BalanceTestCase struct{}

func (t *BalanceTestCase) Name() string {
	return "eth_getBalance"
}

func (t *BalanceTestCase) RequiresContract() bool {
	return false
}

func (t *BalanceTestCase) GetRequests(cfg config.Config) []pkgTypes.Meta {
	return []pkgTypes.Meta{
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      1,
				Method:  "eth_getBalance",
				Params:  []interface{}{cfg.From, "latest"},
			},
			Desc: "Valid account balance request",
		},
		{
			JsonRpcRequest: pkgTypes.JsonRpcRequest{
				JsonRpc: "2.0",
				Id:      2,
				Method:  "eth_getBalance",
				Params:  []interface{}{"0x1234", "latest"},
			},
			Desc: "Invalid account format",
		},
	}
}

func (t *BalanceTestCase) Execute(cfg config.Config) {
	requests := t.GetRequests(cfg)
	jsonrpc.SendReq(requests, cfg)
}

func NewBalanceTestCase() pkgTypes.TestCase {
	return &BalanceTestCase{}
}
