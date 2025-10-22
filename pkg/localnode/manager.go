package localnode

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/jsonrpc"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NodeManager struct {
	config      config.Config
	containerID string
	started     bool
}

func NewNodeManager(cfg config.Config) *NodeManager {
	return &NodeManager{
		config:  cfg,
		started: false,
	}
}

func (nm *NodeManager) StartAndFund() (string, error) {
	if !nm.config.IsLocalNode() {
		return "", nil
	}

	fmt.Printf("Starting %s in Docker dev mode...\n", nm.config.LocalNodeType)

	containerName := fmt.Sprintf("eip-test-%s", nm.config.LocalNodeType)

	checkCmd := exec.Command("docker", "ps", "-aq", "--filter", fmt.Sprintf("name=%s", containerName))
	if output, err := checkCmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		exec.Command("docker", "rm", "-f", containerName).Run()
	}

	var cmd *exec.Cmd
	switch nm.config.LocalNodeType {
	case "geth":
		cmd = exec.Command("docker", "run", "-d",
			"--name", containerName,
			"-p", "8545:8545",
			"ethereum/client-go:latest",
			"--dev",
			"--dev.period", "1",
			"--http",
			"--http.addr", "0.0.0.0",
			"--http.port", "8545",
			"--http.api", "eth,net,web3,debug,personal",
			"--http.corsdomain", "*",
			"--allow-insecure-unlock",
			"--verbosity", "3",
			"--gpo.ignoreprice", "0",
			"--password", "/dev/null", // Empty password for dev mode
		)
	case "besu":
		// Get the absolute path to the genesis.json file
		_, currentFile, _, _ := runtime.Caller(0)
		pkgDir := filepath.Dir(currentFile)
		genesisPath := filepath.Join(pkgDir, "besu", "genesis.json")
		// https://besu.hyperledger.org/public-networks/reference/cli/options#min-priority-fee
		cmd = exec.Command("docker", "run", "-d",
			"--name", containerName,
			"-p", "8545:8545",
			// https://github.com/hyperledger/besu/blob/750580dcca349d22d024cc14a8171b2fa74b505a/config/src/main/resources/dev.json
			"-v", fmt.Sprintf("%s:/genesis.json:ro", genesisPath), // Mount genesis file as read-only
			"hyperledger/besu:latest",
			"--genesis-file=/genesis.json",
			"--miner-enabled",
			"--miner-coinbase=0xfe3b557e8fb62b89f4916b721be55ceb828dbd73",
			"--rpc-http-enabled",
			"--rpc-http-host=0.0.0.0",
			"--rpc-http-port=8545",
			"--rpc-http-api=ETH,NET,WEB3,DEBUG",
			"--rpc-http-cors-origins=*",
			"--host-allowlist=*",
			"--rpc-gas-cap=167700000",
			"--min-gas-price=10",
			"--min-priority-fee=10",
			"--rpc-tx-feecap=100000000000",
			"--logging=DEBUG",
		)
	default:
		return "", fmt.Errorf("unsupported client: %s", nm.config.LocalNodeType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to start: %w\n%s", err, string(output))
	}

	nm.containerID = strings.TrimSpace(string(output))[:12]
	nm.started = true

	fmt.Printf("Container started: %s\n", nm.containerID)

	if err := nm.waitForNode(); err != nil {
		nm.Stop()
		return "", err
	}

	fmt.Println("Node ready")
	devAccount, err := nm.getDevAccount()
	if err != nil {
		nm.Stop()
		return "", fmt.Errorf("failed to get dev account: %w", err)
	}

	if devAccount != "" {
		fmt.Printf("Using Dev account: %s\n", devAccount)
		if err := nm.FundAccount(devAccount, nm.config.From); err != nil {
			return "", fmt.Errorf("failed to fund test account: %w", err)
		}
	}

	return devAccount, nil
}

func (nm *NodeManager) waitForNode() error {
	for i := 0; i < 60; i++ {
		client, err := ethclient.Dial(nm.config.Url)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := client.ChainID(ctx)
			cancel()
			client.Close()
			if err == nil {
				return nil
			}
		}
		if i%5 == 0 && i > 0 {
			fmt.Printf("Waiting... (%d/60)\n", i)
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout")
}

func (nm *NodeManager) getDevAccount() (string, error) {
	request := pkgTypes.JsonRpcRequest{
		JsonRpc: "2.0",
		Method:  "eth_accounts",
		Params:  []interface{}{},
		Id:      1,
	}

	response, err := jsonrpc.SendRawJSONRPCRequest(nm.config.Url, []pkgTypes.JsonRpcRequest{request})
	if err != nil {
		return "", fmt.Errorf("failed to query accounts: %w", err)
	}

	var batchResult []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &batchResult); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(batchResult) == 0 {
		return "", fmt.Errorf("empty response")
	}

	result := batchResult[0]
	if errObj, ok := result["error"]; ok {
		return "", fmt.Errorf("RPC error: %v", errObj)
	}

	// Parse the accounts array
	resultData, ok := result["result"].([]interface{})
	if !ok || len(resultData) == 0 {
		return "", nil
	}

	address, ok := resultData[0].(string)
	if !ok || address == "" || !strings.HasPrefix(address, "0x") {
		return "", fmt.Errorf("invalid address format: %v", resultData[0])
	}

	return address, nil
}

func (nm *NodeManager) Stop() error {
	if !nm.started {
		return nil
	}
	fmt.Printf("Stopping %s...\n", nm.config.LocalNodeType)
	containerName := fmt.Sprintf("eip-test-%s", nm.config.LocalNodeType)
	exec.Command("docker", "rm", "-f", containerName).Run()
	nm.started = false
	return nil
}

func (nm *NodeManager) IsRunning() bool {
	return nm.started
}

// FundAccount funds a target account from the dev account
func (nm *NodeManager) FundAccount(fromAddr, toAddr string) error {
	client, err := ethclient.Dial(nm.config.Url)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer client.Close()

	to := common.HexToAddress(toAddr)

	balance, err := client.BalanceAt(context.Background(), to, nil)
	if err != nil {
		return fmt.Errorf("failed to check balance: %w", err)
	}

	oneEth := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18))
	if balance.Cmp(oneEth) > 0 {
		return nil
	}

	var txHash string
	if nm.config.LocalNodeType == "geth" {
		// Geth dev mode has unlocked accounts, use eth_sendTransaction
		txHash, err = nm.fundAccountUnlocked(fromAddr, toAddr)
		if err != nil {
			return err
		}
		receipt, err := jsonrpc.WaitForTransaction(client, txHash)
		if err != nil {
			return err
		}
		if receipt.Status == 1 {
			balance, err := client.BalanceAt(context.Background(), to, nil)
			if err != nil {
				return fmt.Errorf("failed to verify balance: %w", err)
			}
			fmt.Printf("Test account: %s funded: (balance: %s wei)\n", to.Hex(), balance.String())
		}
	}

	return nil
}

func (nm *NodeManager) fundAccountUnlocked(fromAddr, toAddr string) (string, error) {
	value := "0x56bc75e2d63100000" // 100 ETH in hex

	txParams := map[string]interface{}{
		"from":  fromAddr,
		"to":    toAddr,
		"value": value,
		"gas":   "0x5208", // 21000
	}

	request := pkgTypes.JsonRpcRequest{
		JsonRpc: "2.0",
		Method:  "eth_sendTransaction",
		Params:  []interface{}{txParams},
		Id:      1,
	}

	response, err := jsonrpc.SendRawJSONRPCRequest(nm.config.Url, []pkgTypes.JsonRpcRequest{request})
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txhashes, err := jsonrpc.BatchResponseToTxHashes(response)
	if err != nil {
		return "", fmt.Errorf("failed to parse transaction hash: %w", err)
	}

	return txhashes[0], nil
}
