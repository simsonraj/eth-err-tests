package deployer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/contract"
	pkgTypes "github.com/eth-error-tests/pkg/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Deployer struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	config     config.Config
}

func NewDeployer(cfg config.Config) (*Deployer, error) {
	client, err := ethclient.Dial(cfg.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to client: %w", err)
	}

	chainID := big.NewInt(cfg.ChainID)

	// If no private key is provided (dev mode), we'll use the node to sign
	var privateKey *ecdsa.PrivateKey
	if cfg.PrivateKey != "" {
		privateKey, err = crypto.HexToECDSA(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
	} else {
		return nil, fmt.Errorf("private key is required for deployment, set PRIVATE_KEY environment variable")
	}

	return &Deployer{
		client:     client,
		privateKey: privateKey,
		chainID:    chainID,
		config:     cfg,
	}, nil
}

func (d *Deployer) DeployContract(contractName contract.Name) (*pkgTypes.DeploymentResult, error) {
	fmt.Printf("Deploying contract: %s\n", contractName)

	artifact, err := contract.ArtifactFromContract(contractName)
	if err != nil {
		return &pkgTypes.DeploymentResult{
			ContractName: string(contractName),
			Success:      false,
			Error:        err,
		}, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(d.privateKey, d.chainID)
	if err != nil {
		return &pkgTypes.DeploymentResult{
			ContractName: string(contractName),
			Success:      false,
			Error:        err,
		}, err
	}

	nonce, err := d.client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return &pkgTypes.DeploymentResult{
			ContractName: string(contractName),
			Success:      false,
			Error:        err,
		}, err
	}
	auth.Nonce = big.NewInt(int64(nonce))

	gasPrice, err := d.client.SuggestGasPrice(context.Background())
	if err != nil {
		return &pkgTypes.DeploymentResult{
			ContractName: string(contractName),
			Success:      false,
			Error:        err,
		}, err
	}
	auth.GasPrice = gasPrice
	auth.GasLimit = 3000000 // 3M gas should be enough for most contract deployments

	address, tx, _, err := bind.DeployContract(auth, artifact.Abi, common.FromHex(artifact.Bytecode), d.client)
	if err != nil {
		return &pkgTypes.DeploymentResult{
			ContractName: string(contractName),
			Success:      false,
			Error:        err,
		}, err
	}

	fmt.Printf("Contract %s deployment transaction sent: %s\n", contractName, tx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), d.client, tx)
	if err != nil {
		return &pkgTypes.DeploymentResult{
			ContractName:    string(contractName),
			ContractAddress: address,
			TxHash:          tx.Hash(),
			Success:         false,
			Error:           err,
		}, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		err := fmt.Errorf("deployment transaction failed")
		return &pkgTypes.DeploymentResult{
			ContractName:    string(contractName),
			ContractAddress: address,
			TxHash:          tx.Hash(),
			Success:         false,
			Error:           err,
		}, err
	}

	fmt.Printf("Contract %s deployed successfully at: %s\n", contractName, address.Hex())

	return &pkgTypes.DeploymentResult{
		ContractName:    string(contractName),
		ContractAddress: address,
		TxHash:          tx.Hash(),
		Success:         true,
		Error:           nil,
	}, nil
}

func (d *Deployer) DeploySpecificContracts(contractNames []contract.Name) (map[string]common.Address, []error) {
	deployedContracts := make(map[string]common.Address)
	var errors []error

	for _, name := range contractNames {
		result, err := d.DeployContract(name)
		if err != nil {
			fmt.Printf("Failed to deploy %s: %v\n", name, err)
			errors = append(errors, err)
			continue
		}

		if result.Success {
			deployedContracts[string(name)] = result.ContractAddress
		} else {
			errors = append(errors, result.Error)
		}
	}

	return deployedContracts, errors
}

func (d *Deployer) Close() {
	if d.client != nil {
		d.client.Close()
	}
}
