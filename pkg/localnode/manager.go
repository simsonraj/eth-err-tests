package localnode

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/eth-error-tests/pkg/config"
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

func (nm *NodeManager) Start() (string, error) {
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

// need to fix this instead of curl, it should be an raw RPC call
func (nm *NodeManager) getDevAccount() (string, error) {
	cmd := exec.Command("curl", "-s", "-X", "POST",
		"http://localhost:8545",
		"-H", "Content-Type: application/json",
		"--data", `{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query accounts: %w", err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, `"error"`) {
		return "", fmt.Errorf("RPC error: %s", outputStr)
	}
	start := strings.Index(outputStr, `"result":["`) + len(`"result":["`)
	if start < len(`"result":["`) {
		return "", fmt.Errorf("no accounts found in response: %s", outputStr)
	}

	end := strings.Index(outputStr[start:], `"`)
	if end < 0 {
		return "", fmt.Errorf("failed to parse account address from: %s", outputStr)
	}

	address := outputStr[start : start+end]
	if address == "" || !strings.HasPrefix(address, "0x") {
		return "", fmt.Errorf("invalid address format: %s", address)
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
