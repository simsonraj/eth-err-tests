package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/runner"
	"github.com/spf13/cobra"
)

func Report(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	outputFile, err := os.Create(filename + ".csv")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			fmt.Printf("Error closing output file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	scenarioRegexp := regexp.MustCompile(`^Scenario: (.+?)\s+-\s+Request:`)
	requestRegexp := regexp.MustCompile(`(?s)Request: (.+)`)
	responseRegexp := regexp.MustCompile(`(?s)^Response: (.+)$`)
	methodRegex := regexp.MustCompile(`"method":"([^"]+)"`)
	err = writer.Write([]string{"Method", "scenario", "Response", "Request"})
	if err != nil {
		panic(err)
	}
	var scenario, request, response, method string
	for scanner.Scan() {
		line := scanner.Text()
		if matches := scenarioRegexp.FindStringSubmatch(line); len(matches) > 0 {
			scenario = matches[1]
			fmt.Println(scenario)
		}
		if matches := requestRegexp.FindStringSubmatch(line); len(matches) > 0 {
			request = matches[1]
			if newmatches := methodRegex.FindStringSubmatch(request); len(newmatches) > 0 {
				method = newmatches[1]
			}
		} else if matches := responseRegexp.FindStringSubmatch(line); len(matches) > 0 {
			response = matches[1]
			err = writer.Write([]string{method, scenario, response, request})
			if err != nil {
				panic(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}
}

var (
	env   string
	tests string
)

var rootCmd = &cobra.Command{
	Use: "eth-err-tests",
	Long: `Ethereum RPC Client Tester
	Docker Clients: besu-local, geth-local, reth-local, nethermind-local, erigon-local
	Remote Networks: sepolia, zkevm`,
	Example: ` #Test on local geth via Docker
  eth-err-tests --env geth-local

  # Run specific tests on zkEVM
  eth-err-tests --env zkevm --tests eth_call,eth_estimateGas`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(env)
		if err != nil {
			fmt.Printf("Error: Invalid environment '%s': %v\n", env, err)
			os.Exit(1)
		}

		fmt.Printf("Testing Network: %s\n", cfg.Network)
		fmt.Printf("RPC URL: %s\n", cfg.Url)
		fmt.Println("=======================================================")
		fmt.Println()

		testRunner, err := runner.NewTestRunner(cfg)
		if err != nil {
			fmt.Printf("Error creating test runner: %v\n", err)
			os.Exit(1)
		}
		defer testRunner.Cleanup()

		var testNames []string
		if tests != "" {
			testNames = strings.Split(tests, ",")
			for i := range testNames {
				testNames[i] = strings.TrimSpace(testNames[i])
			}
		}

		if err := testRunner.RunWithAutoDeployment(testNames); err != nil {
			fmt.Printf("Error running tests: %v\n", err)
			os.Exit(1)
		}
	},
}

var reportCmd = &cobra.Command{
	Use:     "report [logfile]",
	Long:    "Generate a CSV report from a test log file",
	Example: `eth-err-tests report reports/geth-local.log`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logFile := args[0]
		fmt.Println("Generating report from log file:", logFile)
		Report(logFile)
	},
}

func init() {
	// Root command
	rootCmd.Flags().StringVarP(&env, "env", "e", "", "Network/client to test (required)")
	rootCmd.Flags().StringVarP(&tests, "tests", "t", "", "Comma-separated list of tests to run (e.g., eth_getBalance,eth_getCode,eth_call,eth_estimateGas,eth_sendRawTransaction)")
	if err := rootCmd.MarkFlagRequired("env"); err != nil {
		panic(err)
	}

	// report command
	rootCmd.AddCommand(reportCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
