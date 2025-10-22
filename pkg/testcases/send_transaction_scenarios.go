package testcases

import (
	"context"
	"fmt"
	"math/big"

	"github.com/eth-error-tests/pkg/config"
	txbuilder "github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetScenarios(cfg config.Config) []pkgTypes.Scenario {
	return []pkgTypes.Scenario{
		{
			ID:     1,
			Desc:   "Proper request",
			Method: "eth_sendRawTransaction",
		},
		{
			ID:     3,
			Desc:   "UseInvalidFunction",
			Method: "eth_wrongSendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.InvalidFunctionSigModifier("invalidFunction(uint256)", 20),
			},
		},
		{
			ID:     4,
			Desc:   "NONCE_TOO_LOW",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.NonceModifier(0, nil), // Set nonce to 0
			},
		},
		{
			ID:     5,
			Desc:   "NONCE_TOO_HIGH", // DOesnt work
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.NonceModifier(0, func(current uint64) uint64 { return current + 10000 }),
			},
		},
		{
			ID:     9,
			Desc:   "OVERSIZED_DATA",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.DataSizeModifier(1024 * 1024), // 1 MB
			},
		},
		{
			ID:     10,
			Desc:   "BLOCK_GAS_LIMIT_EXCEEDED",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasLimitModifier(0, func(current uint64) uint64 {
					return 46_000_000
				}), // Set gas limit to 46_000_000
			},
		},
		{
			ID:     10,
			Desc:   "TRANSACTION_GAS_LIMIT_EXCEEDED: GasLimitTooHigh",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasLimitModifier(0, func(current uint64) uint64 {
					return 16_777_216 + 1
				}), // Set gas limit to max transaction gas limit + 1
			},
		},
		{
			ID:     11,
			Desc:   "GAS_PRICE_TOO_LOW-Legacy",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasPriceModifier(big.NewInt(0), nil),
			},
		},
		{
			ID:     11,
			Desc:   "GAS_PRICE_TOO_LOW-Dynamic",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasTipCapModifier(big.NewInt(0), nil),
				txbuilder.GasFeeCapModifier(big.NewInt(0), nil),
			},
		},
		{
			ID:     12,
			Desc:   "FEE_CAP_EXCEEDED",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasLimitModifier(16_000_000, nil),
				txbuilder.GasPriceModifier(big.NewInt(200000000000), nil),
			},
		},

		{
			ID:     13,
			Desc:   "GAS_TOO_LOW - Intrinsic gas too low",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasLimitModifier(20000, nil), // Set gas limit to 20000 (below 21000 intrinsic gas)
			},
		},
		{
			ID:     14,
			Desc:   "OUT_OF_GAS - Transaction runs out of gas",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasLimitModifier(0, func(current uint64) uint64 {
					return 21204 // return only intricsic gas
				}),
			},
		},
		{
			ID:     17,
			Desc:   "TipAboveFeeCap - max priority fee per gas higher than max fee per gas",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.GasFeeCapModifier(big.NewInt(5), nil),
			},
		},
		{
			ID:     20,
			Desc:   "INSUFFICIENT_FUNDS - Not enough funds for gas * price + value",
			Method: "eth_sendRawTransaction",
			Modifiers: []pkgTypes.Modifier{
				txbuilder.ValueFromBalanceModifier(1000000000000000000), // balance + 1 ETH
			},
		},
		{
			ID:       21,
			Desc:     "REPLACEMENT_TRANSACTION_UNDERPRICED - Replacement without price bump",
			Method:   "eth_sendRawTransaction",
			UseBatch: true,
			PreSend: func(ctx context.Context, client *ethclient.Client, cfg config.Config, params *pkgTypes.TxParams) (string, error) {
				var firstTx *types.Transaction
				if params.IsDynamic {
					firstTx = types.NewTx(&types.DynamicFeeTx{
						ChainID:   big.NewInt(params.ChainID),
						Nonce:     params.Nonce,
						GasTipCap: params.GasTipCap,
						GasFeeCap: params.GasFeeCap,
						Gas:       params.Gas,
						To:        params.To,
						Value:     big.NewInt(1000), // Different value to make it a different transaction
						Data:      params.Data,
					})
				} else {
					firstTx = types.NewTransaction(params.Nonce, *params.To, big.NewInt(1000), params.Gas, params.GasPrice, params.Data)
				}

				signedFirstTx, err := txbuilder.SignTransaction(firstTx, params)
				if err != nil {
					return "", fmt.Errorf("error signing first transaction: %w", err)
				}

				encodedFirstTx, err := signedFirstTx.MarshalBinary()
				if err != nil {
					return "", fmt.Errorf("error encoding first transaction: %w", err)
				}

				rawFirstTx := "0x" + common.Bytes2Hex(encodedFirstTx)
				return rawFirstTx, nil
			},
		},
		{
			ID:       22,
			Desc:     "ALREADY_KNOWN ",
			Method:   "eth_sendRawTransaction",
			UseBatch: true,
			PreSend: func(ctx context.Context, client *ethclient.Client, cfg config.Config, params *pkgTypes.TxParams) (string, error) {
				var firstTx *types.Transaction
				if params.IsDynamic {
					firstTx = types.NewTx(&types.DynamicFeeTx{
						ChainID:   big.NewInt(params.ChainID),
						Nonce:     params.Nonce,
						GasTipCap: params.GasTipCap,
						GasFeeCap: params.GasFeeCap,
						Gas:       params.Gas,
						To:        params.To,
						Value:     params.Value,
						Data:      params.Data,
					})
				} else {
					firstTx = types.NewTransaction(params.Nonce, *params.To, big.NewInt(1000), params.Gas, params.GasPrice, params.Data)
				}

				signedFirstTx, err := txbuilder.SignTransaction(firstTx, params)
				if err != nil {
					return "", fmt.Errorf("error signing first transaction: %w", err)
				}

				encodedFirstTx, err := signedFirstTx.MarshalBinary()
				if err != nil {
					return "", fmt.Errorf("error encoding first transaction: %w", err)
				}

				rawFirstTx := "0x" + common.Bytes2Hex(encodedFirstTx)
				return rawFirstTx, nil
			},
		},

		/*
			// need to write revert opcode & invalid opcode in contracts
				{
					ID:     17,
					Desc:   "FeeCapTooLow - max fee per gas less than block base fee", // Fix this by overriding genesis to have high base fee
					Method: "eth_sendRawTransaction",
					Modifiers: []pkgTypes.Modifier{
						GasPriceModifier(big.NewInt(5), nil),
						// GasFeeCapModifier(big.NewInt(10), nil),
						// GasTipCapModifier(big.NewInt(1), nil),
					},
				},
							{
								ID:     19,
								Desc:   "INVALID_MAX_FEE_PER_GAS - max fee per gas higher than 2^256-1", // Wont work because the FEE CAP is set to 1 ETH
								Method: "eth_sendRawTransaction",
								Modifiers: []pkgTypes.Modifier{
									GasFeeCapModifier(new(big.Int).Lsh(big.NewInt(1), 255), nil), // Set to 2^255
								},
							},
							{
							ID:     18,
							Desc:   "INVALID_MAX_PRIORITY_FEE_PER_GAS - max priority fee per gas higher than 2^256-1", // Wont work because the FEE CAP is set to 1 ETH
							Method: "eth_sendRawTransaction",
							Modifiers: []pkgTypes.Modifier{
								GasTipCapModifier(new(big.Int).Lsh(big.NewInt(1), 255), nil), // Set to 2^255
								GasFeeCapModifier(new(big.Int).Lsh(big.NewInt(1), 255), nil), // Set to 2^255
							},
						},
								{
									ID:     16,
									Desc:   "GAS_OVERFLOW - Gas overflow error", // TODO FIX this by creating a txn that overflows gas calculation
									Method: "eth_sendRawTransaction",
									Modifiers: []pkgTypes.Modifier{
										GasLimitModifier(0xFFFFFFFFFFFFFF, nil), // Set to large value that triggers overflow
									},
								},
									{
										ID:     7,
										Desc:   "UseInvalidAccount",
										Method: "eth_sendRawTransaction",
										Modifiers: []pkgTypes.Modifier{
											PrivateKeyModifier("f7d9eb9afde6a5da1e7257e0d1c1c7b7f0e5476a8f6bfc9f1c2e076fcff6a2a6"),
										},
									},
									{
										ID:     8,
										Desc:   "UseInvalidContract", // Doesnt work
										Method: "eth_sendRawTransaction",
										Modifiers: []pkgTypes.Modifier{
											ToAddressModifier("0x0000000000000000000000000000000000000000"),
										},
									}, */
	}
}
