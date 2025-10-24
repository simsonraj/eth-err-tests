package config

import (
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// Config holds the configuration for connecting to an Ethereum client
type Config struct {
	Network           string
	Url               string
	From              string // Will be updated with derived address from PrivateKey
	ToContract        string // Will be updated with deployed contract address
	DeployedContracts map[string]common.Address
	PrivateKey        string
	ChainID           int64
	InvalidContract   string
	LocalNodeType     string // Type of local node: "besu", "geth", "reth", etc.
}

var (
	zkSyncConfig = Config{
		Network:         "zksync",
		Url:             "https://zksync2-testnet.zksync.dev",
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
		ChainID:         280,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
	}

	sepoliaConfig = Config{
		Network:    "Sepolia",
		Url:        "https://ethereum-sepolia-rpc.publicnode.com",
		PrivateKey: os.Getenv("PRIVATE_KEY"),
		DeployedContracts: map[string]common.Address{
			"storage":    common.HexToAddress("0xfB1fa32605b1Cd1B91d1B80CCC7dde8EDab643D3"),
			"opcodes":    common.HexToAddress("0xBeDBEdeB6362681Ad8d1f87FC600418db3A20521"),
			"testkeccak": common.HexToAddress("0xF603Da4416330A1715ADb9B5407505573C4FD8c7"),
		},
		ToContract:      "0xfB1fa32605b1Cd1B91d1B80CCC7dde8EDab643D3",
		ChainID:         11155111,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
	}

	zkEVMConfig = Config{
		Network:         "zkEVM",
		Url:             "https://rpc.public.zkevm-test.net",
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
		ChainID:         1442,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
	}

	besuLocalConfig = Config{
		Network:         "besu-local",
		Url:             "http://localhost:8545",
		PrivateKey:      os.Getenv("PRIVATE_KEY"), // Get private key from https://besu.hyperledger.org/private-networks/reference/accounts-for-testing
		ChainID:         1337,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
		LocalNodeType:   "besu",
	}

	// Local Geth node configuration
	gethLocalConfig = Config{
		Network:         "geth-local",
		Url:             "http://localhost:8545",
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
		ChainID:         1337,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
		LocalNodeType:   "geth",
	}

	// Local Reth node configuration
	rethLocalConfig = Config{
		Network:         "reth-local",
		Url:             "http://localhost:8545",
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
		ChainID:         1337,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
		LocalNodeType:   "reth",
	}

	// Local Nethermind node configuration
	nethermindLocalConfig = Config{
		Network:         "nethermind-local",
		Url:             "http://localhost:8545",
		PrivateKey:      os.Getenv("PRIVATE_KEY"), // Dev mode test private key
		ChainID:         1337,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
		LocalNodeType:   "nethermind",
	}

	// Local Erigon node configuration
	erigonLocalConfig = Config{
		Network:         "erigon-local",
		Url:             "http://localhost:8545",
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
		ChainID:         1337,
		InvalidContract: "0x0baEAd25fe0346B76C73e84c083bb503c14309F1",
		LocalNodeType:   "erigon",
	}
)

func GetConfig(env string) (Config, error) {
	switch env {
	case "zksync":
		return zkSyncConfig, nil
	case "sepolia":
		return sepoliaConfig, nil
	case "zkevm":
		return zkEVMConfig, nil
	case "besu-local", "besu":
		return besuLocalConfig, nil
	case "geth-local", "geth":
		return gethLocalConfig, nil
	case "reth-local", "reth":
		return rethLocalConfig, nil
	case "nethermind-local", "nethermind":
		return nethermindLocalConfig, nil
	case "erigon-local", "erigon":
		return erigonLocalConfig, nil
	default:
		return Config{}, errors.New("invalid environment: " + env)
	}
}

func (c *Config) IsLocalNode() bool {
	return c.LocalNodeType != ""
}
