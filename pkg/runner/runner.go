package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/contract"
	"github.com/eth-error-tests/pkg/deployer"
	"github.com/eth-error-tests/pkg/localnode"
	"github.com/eth-error-tests/pkg/testcases"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TestRunner struct {
	config            config.Config
	deployedContracts map[string]common.Address
	deployer          *deployer.Deployer
	nodeManager       *localnode.NodeManager
}

func NewTestRunner(cfg config.Config) (*TestRunner, error) {
	if cfg.PrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		cfg.From = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	} else {
		return nil, fmt.Errorf("private key is required for deployment, set PRIVATE_KEY environment variable")
	}

	return &TestRunner{
		config:            cfg,
		deployedContracts: make(map[string]common.Address),
		nodeManager:       localnode.NewNodeManager(cfg),
	}, nil
}

func (r *TestRunner) DeployContracts() error {
	fmt.Printf("Deploying contracts to %s (%s)\n", r.config.Network, r.config.Url)

	d, err := deployer.NewDeployer(r.config)
	if err != nil {
		return fmt.Errorf("failed to create deployer: %w", err)
	}
	r.deployer = d

	// Extend this to deploy all contracts if needed
	contractsToDepl := []contract.Name{
		contract.Storage,
	}

	deployedContracts, errors := r.deployer.DeploySpecificContracts(contractsToDepl)
	if len(errors) > 0 {
		fmt.Printf("Warning: Some contracts failed to deploy: %v\n", errors)
	}

	r.deployedContracts = deployedContracts

	// Update the config with deployed contract addresses
	if storageAddr, ok := deployedContracts["storage"]; ok {
		r.config.ToContract = storageAddr.Hex()
	} else {
		return fmt.Errorf("failed to deploy Storage contract")
	}

	return nil
}

func (r *TestRunner) RunAllTests() error {
	testSuites := testcases.GetAllTestSuites()

	for _, suite := range testSuites {
		if err := r.RunTestSuite(suite); err != nil {
			fmt.Printf("Error running suite %s: %v\n", suite.Name, err)
		}
	}

	return nil
}

func (r *TestRunner) RunTestSuite(suite pkgTypes.TestSuite) error {
	fmt.Printf("Running Test Suite: %s\n", suite.Name)
	fmt.Printf("Description: %s\n", suite.Description)

	/* if suite.RequiresContracts && len(r.deployedContracts) == 0 {
		if err := r.DeployContracts(); err != nil {
			return fmt.Errorf("failed to deploy contracts: %w", err)
		}
	} */

	for i, testCase := range suite.TestCases {
		fmt.Printf("\n[%d/%d] Running Test: %s\n", i+1, len(suite.TestCases), testCase.Name())
		fmt.Println("-------------------------------------------------------")

		testCase.Execute(r.config)

		fmt.Println("-------------------------------------------------------")
	}

	fmt.Println()
	return nil
}

func (r *TestRunner) RunSpecificTest(testName string) error {
	testCase := testcases.GetTestCaseByName(testName)
	if testCase == nil {
		return fmt.Errorf("test case '%s' not found", testName)
	}

	fmt.Println("=======================================================")
	fmt.Printf("Running Test: %s\n", testCase.Name())
	fmt.Println("=======================================================")

	if testCase.RequiresContract() && len(r.deployedContracts) == 0 {
		if err := r.DeployContracts(); err != nil {
			return fmt.Errorf("failed to deploy contracts: %w", err)
		}
	}

	startTime := time.Now()
	testCase.Execute(r.config)
	duration := time.Since(startTime)

	fmt.Printf("\nTest completed in %v\n", duration)
	fmt.Println("=======================================================")
	return nil
}

func (r *TestRunner) RunWithAutoDeployment(testNames []string) error {
	if r.config.IsLocalNode() {
		devAccount, err := r.nodeManager.Start()
		if err != nil {
			return fmt.Errorf("failed to start local node: %w", err)
		}

		if devAccount != "" {
			fmt.Printf("Using Dev account: %s\n", devAccount)
			if err := r.fundAccount(devAccount, r.config.From); err != nil {
				return fmt.Errorf("failed to fund test account: %w", err)
			}
		}
	}
	if r.config.ToContract == "" {
		if err := r.DeployContracts(); err != nil {
			return fmt.Errorf("failed to deploy contracts: %w", err)
		}
	}

	if len(testNames) == 0 {
		return r.RunAllTests()
	}

	for _, testName := range testNames {
		if err := r.RunSpecificTest(testName); err != nil {
			fmt.Printf("Error running test %s: %v\n", testName, err)
		}
	}

	return nil
}

func (r *TestRunner) fundAccount(fromAddr, toAddr string) error {
	client, err := ethclient.Dial(r.config.Url)
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

	value := "0x56bc75e2d63100000" // 100 ETH in hex

	txParams := map[string]interface{}{
		"from":  fromAddr,
		"to":    toAddr,
		"value": value,
		"gas":   "0x5208", // 21000
	}

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_sendTransaction",
		"params":  []interface{}{txParams},
		"id":      1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(r.config.Url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if errObj, ok := result["error"]; ok {
		return fmt.Errorf("RPC error: %v", errObj)
	}

	txHash, ok := result["result"].(string)
	if !ok {
		return fmt.Errorf("unexpected response format: %v", result)
	}

	for i := 0; i < 30; i++ {
		receiptReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "eth_getTransactionReceipt",
			"params":  []interface{}{txHash},
			"id":      1,
		}

		receiptJSON, err := json.Marshal(receiptReq)
		if err != nil {
			return fmt.Errorf("failed to marshal receipt request: %w", err)
		}

		receiptResp, err := http.Post(r.config.Url, "application/json", bytes.NewBuffer(receiptJSON))
		if err != nil {
			return fmt.Errorf("failed to get receipt: %w", err)
		}

		receiptBody, err := ioutil.ReadAll(receiptResp.Body)
		receiptResp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read receipt response: %w", err)
		}

		var receiptResult map[string]interface{}
		if err := json.Unmarshal(receiptBody, &receiptResult); err != nil {
			return fmt.Errorf("failed to parse receipt response: %w", err)
		}

		if receipt, ok := receiptResult["result"].(map[string]interface{}); ok && receipt != nil {
			if status, ok := receipt["status"].(string); ok && status == "0x1" {
				balance, err := client.BalanceAt(context.Background(), to, nil)
				if err != nil {
					return fmt.Errorf("failed to verify balance: %w", err)
				}
				fmt.Printf("Test account: %s funded: (balance: %s wei)\n", to.Hex(), balance.String())
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for funding transaction to be mined")
}

func (r *TestRunner) Cleanup() {
	if r.deployer != nil {
		r.deployer.Close()
	}

	if r.nodeManager != nil && r.nodeManager.IsRunning() {
		if err := r.nodeManager.Stop(); err != nil {
			fmt.Printf("Warning: failed to stop local node: %v\n", err)
		}
	}
}

func (r *TestRunner) GetDeployedContracts() map[string]common.Address {
	return r.deployedContracts
}
