package jsonrpc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/eth-error-tests/pkg/config"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewTxParamsFromDefaults(ctx context.Context, client *ethclient.Client, cfg config.Config, privateKey *ecdsa.PrivateKey, toAddress common.Address, input []byte) (*pkgTypes.TxParams, error) {
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting nonce: %w", err)
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddress,
		To:   &toAddress,
		Data: input,
	})
	if err != nil {
		gasLimit = 100000
	}

	// Get suggested gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting gas price: %w", err)
	}

	// Check if network supports EIP-1559
	header, err := client.HeaderByNumber(ctx, nil)
	isDynamic := false
	var gasTipCap, gasFeeCap *big.Int

	if err == nil && header.BaseFee != nil {
		// Network supports EIP-1559
		isDynamic = true
		gasTipCap = big.NewInt(2000000000)                                    // 2 gwei
		gasFeeCap = new(big.Int).Add(header.BaseFee, big.NewInt(10000000000)) // baseFee + 10 gwei
	}

	return &pkgTypes.TxParams{
		Nonce:       nonce,
		To:          &toAddress,
		Value:       big.NewInt(0),
		Data:        input,
		Gas:         gasLimit,
		GasPrice:    gasPrice,
		GasTipCap:   gasTipCap,
		GasFeeCap:   gasFeeCap,
		IsDynamic:   isDynamic,
		ChainID:     cfg.ChainID,
		PrivateKey:  privateKey,
		FromAddress: fromAddress,
	}, nil
}

func BuildTransaction(params *pkgTypes.TxParams) *types.Transaction {
	if params.IsDynamic && params.GasTipCap != nil && params.GasFeeCap != nil {
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(params.ChainID),
			Nonce:     params.Nonce,
			GasTipCap: params.GasTipCap,
			GasFeeCap: params.GasFeeCap,
			Gas:       params.Gas,
			To:        params.To,
			Value:     params.Value,
			Data:      params.Data,
		})
	}
	return types.NewTransaction(params.Nonce, *params.To, params.Value, params.Gas, params.GasPrice, params.Data)
}

func SignTransaction(tx *types.Transaction, params *pkgTypes.TxParams) (*types.Transaction, error) {
	var signer types.Signer
	switch tx.Type() {
	case types.DynamicFeeTxType: // 0x02 (EIP-1559)
		signer = types.NewLondonSigner(big.NewInt(params.ChainID))
	default: // 0x00 (Legacy)
		signer = types.NewEIP155Signer(big.NewInt(params.ChainID))
	}

	return types.SignTx(tx, signer, params.PrivateKey)
}

// --- Generic Modifiers with Transformation Functions ---
func GasLimitModifier(value uint64, transform func(current uint64) uint64) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		if transform != nil {
			params.Gas = transform(params.Gas)
		} else {
			params.Gas = value
		}
		return nil
	}
}

func GasPriceModifier(value *big.Int, transform func(current *big.Int) *big.Int) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		// if !params.IsDynamic {
		if transform != nil {
			params.GasPrice = transform(params.GasPrice)
			params.IsDynamic = false
		} else {
			params.GasPrice = new(big.Int).Set(value)
			params.IsDynamic = false
		}
		// }
		return nil
	}
}

func GasTipCapModifier(value *big.Int, transform func(current *big.Int) *big.Int) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		if params.IsDynamic {
			if transform != nil {
				params.GasTipCap = transform(params.GasTipCap)
			} else {
				params.GasTipCap = new(big.Int).Set(value)
			}
		}
		return nil
	}
}

func GasFeeCapModifier(value *big.Int, transform func(current *big.Int) *big.Int) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		if params.IsDynamic {
			if transform != nil {
				params.GasFeeCap = transform(params.GasFeeCap)
			} else {
				params.GasFeeCap = new(big.Int).Set(value)
			}
		}
		return nil
	}
}

func ValueModifier(value *big.Int) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		params.Value = new(big.Int).Set(value)
		return nil
	}
}

func ValueFromBalanceModifier(offsetWei int64) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		balance, err := client.BalanceAt(ctx, params.FromAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting balance: %w", err)
		}
		params.Value = new(big.Int).Add(balance, big.NewInt(offsetWei))
		return nil
	}
}

func NonceModifier(value uint64, transform func(current uint64) uint64) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		if transform != nil {
			params.Nonce = transform(params.Nonce)
		} else {
			params.Nonce = value
		}
		return nil
	}
}

func DataModifier(data []byte) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		params.Data = make([]byte, len(data))
		copy(params.Data, data)
		return nil
	}
}

func DataSizeModifier(sizeBytes int) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		params.Data = make([]byte, sizeBytes)
		return nil
	}
}

func PrivateKeyModifier(privateKeyHex string) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return fmt.Errorf("error loading private key: %w", err)
		}
		params.PrivateKey = privateKey
		params.FromAddress = crypto.PubkeyToAddress(privateKey.PublicKey)
		// Re-fetch nonce for the new account
		nonce, err := client.PendingNonceAt(ctx, params.FromAddress)
		if err != nil {
			return fmt.Errorf("error getting nonce for account: %w", err)
		}
		params.Nonce = nonce
		return nil
	}
}

func ToAddressModifier(address string) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		addr := common.HexToAddress(address)
		params.To = &addr
		return nil
	}
}

func InvalidFunctionSigModifier(functionSig string, argValue uint64) pkgTypes.Modifier {
	return func(ctx context.Context, client *ethclient.Client, params *pkgTypes.TxParams) error {
		sig := crypto.Keccak256([]byte(functionSig))[:4]
		arg := common.LeftPadBytes(new(big.Int).SetUint64(argValue).Bytes(), 32)
		params.Data = append(sig, arg...)
		return nil
	}
}
