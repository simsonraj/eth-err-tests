package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/eth-error-tests/pkg/config"
	"github.com/eth-error-tests/pkg/runner"
)

func Report(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	outputFile, err := os.Create(filename + ".csv")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

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

func printUsage() {
	fmt.Println("Ethereum RPC Client Tester")
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ./main --env=<network> [--tests=<test_names>] [--report=<logfile>]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --env        : Network/client to test (required)")
	fmt.Println("                 Docker Clients: besu-local, geth-local, reth-local,")
	fmt.Println("                                 nethermind-local, erigon-local")
	fmt.Println("                 Remote Networks: zksync, zkevm")
	fmt.Println("  --tests      : Comma-separated list of specific tests to run (optional)")
	fmt.Println("                 Available: eth_getBalance, eth_getCode, eth_call,")
	fmt.Println("                            eth_estimateGas, eth_sendRawTransaction")
	fmt.Println("                 If not specified, all tests will be run")
	fmt.Println("  --report     : Generate CSV report from log file (optional)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Test on local Besu via Docker")
	fmt.Println("  ./main --env=besu-local")
	fmt.Println()
	fmt.Println()
	fmt.Println("  # Run specific tests on zkEVM")
	fmt.Println("  ./main --env=zkevm --tests=eth_call,eth_estimateGas")
	fmt.Println()
	fmt.Println("  # Generate report from log file")
	fmt.Println("  ./main --report=zksync.log")
	fmt.Println()
}

func run() int {
	env := flag.String("env", "", "Network/client to test (e.g., zksync, zkevm)")
	testsFlag := flag.String("tests", "", "Comma-separated list of tests to run (optional)")
	logFile := flag.String("report", "", "Generate CSV report from log file")
	help := flag.Bool("help", false, "Show usage information")
	flag.Parse()

	if *help || (flag.NFlag() == 0) {
		printUsage()
		return 0
	}

	if *logFile != "" {
		fmt.Println("Generating report from log file:", *logFile)
		Report(*logFile)
		return 0
	}

	if *env == "" {
		fmt.Println("Error: --env flag is required")
		fmt.Println()
		printUsage()
		return 1
	}

	cfg, err := config.GetConfig(*env)
	if err != nil {
		fmt.Printf("Error: Invalid environment '%s': %v\n", *env, err)
		fmt.Println()
		printUsage()
		return 1
	}

	fmt.Printf("Testing Network: %s\n", cfg.Network)
	fmt.Printf("RPC URL: %s\n", cfg.Url)
	fmt.Println("=======================================================")
	fmt.Println()

	testRunner, err := runner.NewTestRunner(cfg)
	if err != nil {
		fmt.Printf("Error creating test runner: %v\n", err)
		return 1
	}
	defer testRunner.Cleanup()

	var testNames []string
	if *testsFlag != "" {
		testNames = strings.Split(*testsFlag, ",")
		for i := range testNames {
			testNames[i] = strings.TrimSpace(testNames[i])
		}
	}

	if err := testRunner.RunWithAutoDeployment(testNames); err != nil {
		fmt.Printf("Error running tests: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
