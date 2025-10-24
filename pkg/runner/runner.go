package runner

import (
	"fmt"
	"time"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/contract"
	"github.com/eth-error-tests/pkg/deployer"
	"github.com/eth-error-tests/pkg/localnode"
	"github.com/eth-error-tests/pkg/testcases"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
		contract.OpCodes,
		contract.TestKeccak,
	}

	deployedContracts, errors := r.deployer.DeploySpecificContracts(contractsToDepl)
	if len(errors) > 0 {
		fmt.Printf("Warning: Some contracts failed to deploy: %v\n", errors)
	}

	r.deployedContracts = deployedContracts
	r.config.DeployedContracts = deployedContracts

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
		_, err := r.nodeManager.StartAndFund()
		if err != nil {
			return fmt.Errorf("failed to start local node: %w", err)
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
