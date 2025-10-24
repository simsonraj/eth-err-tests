package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/contract"
	"github.com/eth-error-tests/pkg/types"
)

const (
	CONTRACT_ABI = "[{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

func SendRawJSONRPCRequest(url string, requestBody []types.JsonRpcRequest) (string, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func SendReq(requests []types.Meta, cfg config.Config) {
	for _, request := range requests {
		r, err := json.Marshal(request.JsonRpcRequest)
		if err != nil {
			fmt.Println(err)
			return
		}
		reqStr := string(r)
		if len(reqStr) > 1000 {
			reqStr = reqStr[:1000] + "..."
		}
		fmt.Println("Scenario:", request.Desc, " - Request:", reqStr)
		// fmt.Println("Request:", string(r))
		response, err := SendRawJSONRPCRequest(cfg.Url, []types.JsonRpcRequest{request.JsonRpcRequest})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		var data interface{}
		err = json.Unmarshal([]byte(response), &data)
		if err != nil {
			panic(err)
		}

		compactJSON, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}

		fmt.Println("Response:", string(compactJSON))
	}
}

func SendTransaction(ctx context.Context, client *ethclient.Client, cfg config.Config, scenario types.Scenario) error {
	// 1. Load default private key and addresses
	if cfg.PrivateKey == "" {
		return fmt.Errorf("private key is not set in config")
	}
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("error loading private key: %w", err)
	}

	toAddress := common.HexToAddress(cfg.ToContract)

	// 2. Build default input data
	input, err := contract.BuildInput(contract.Storage, "store", new(big.Int).SetUint64(20))
	if err != nil {
		return fmt.Errorf("error building input: %w", err)
	}

	// 3. Build default transaction parameters
	params, err := NewTxParamsFromDefaults(ctx, client, cfg, privateKey, toAddress, input)
	if err != nil {
		return fmt.Errorf("error building tx params: %w", err)
	}

	// 4. Apply modifiers
	for _, modifier := range scenario.Modifiers {
		if err := modifier(ctx, client, params); err != nil {
			return fmt.Errorf("error applying modifier: %w", err)
		}
	}

	// 5. Execute pre-send hook (for replacement scenarios)
	var batchTx string
	var presendErr error
	if scenario.PreSend != nil {
		batchTx, presendErr = scenario.PreSend(ctx, client, cfg, params)
	}

	// 6. Build transaction
	tx := BuildTransaction(params)
	// 7. Sign transaction
	signedTx, err := SignTransaction(tx, params)
	if err != nil {
		return fmt.Errorf("error signing transaction: %w", err)
	}

	// 9. Encode transaction
	encodedTx, err := signedTx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error encoding transaction: %w", err)
	}

	rawTx := "0x" + common.Bytes2Hex(encodedTx)

	// 10. Create JSON-RPC request
	request := []types.JsonRpcRequest{
		{
			JsonRpc: "2.0",
			Id:      scenario.ID,
			Method:  scenario.Method,
			Params:  []interface{}{rawTx},
		},
	}

	// 11. Handle batch requests if needed
	if scenario.UseBatch && batchTx != "" {
		// Send both transactions in a batch
		request = append(request, types.JsonRpcRequest{
			JsonRpc: "2.0",
			Id:      scenario.ID,
			Method:  scenario.Method,
			Params:  []interface{}{batchTx},
		},
		)
	}

	r, _ := json.Marshal(request)
	reqStr := string(r)
	// Truncate long requests for logging
	if len(reqStr) > 1000 {
		reqStr = reqStr[:1000] + "..."
	}
	fmt.Println("Scenario:", scenario.Desc, " - Request:", reqStr)

	// 12. Send transaction
	response, err := SendRawJSONRPCRequest(cfg.Url, request)
	if err != nil {
		fmt.Println("Error:", err)
		return nil // Continue to next scenario
	}

	// 13. Print response
	var data interface{}
	var printResp string
	if err := json.Unmarshal([]byte(response), &data); err == nil {
		compactJSON, _ := json.Marshal(data)
		printResp = string(compactJSON)
	} else {
		printResp = response
	}
	if presendErr != nil {
		printResp += fmt.Sprintf(", PreSend Error: %v", presendErr)
	}

	fmt.Println("Response:", printResp)
	hashes, err := BatchResponseToTxHashes(response)
	if err != nil {
		return fmt.Errorf("failed to parse transaction hashes from response: %w", err)
	}

	if len(hashes) != 0 {
		WaitForTransaction(client, hashes[0])
	}

	return nil
}

func WaitForTransaction(client *ethclient.Client, txHash string) (*gethTypes.Receipt, error) {
	tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	var receipt *gethTypes.Receipt
	if isPending {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		receipt, err = bind.WaitMined(ctx, client, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for transaction mining: %w", err)
		}
	}

	return receipt, nil
}

func BatchResponseToTxHashes(response string) ([]string, error) {
	var batchResult []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &batchResult); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}

	txHashes := make([]string, 0, len(batchResult))
	for _, res := range batchResult {
		if result, ok := res["result"].(string); ok {
			txHashes = append(txHashes, result)
		}
	}

	return txHashes, nil
}
