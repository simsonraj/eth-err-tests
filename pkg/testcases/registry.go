package testcases

import (
	pkgTypes "github.com/eth-error-tests/pkg/types"
)

func GetAllTestSuites() []pkgTypes.TestSuite {
	return []pkgTypes.TestSuite{
		/* {
			Name:              "Basic RPC Tests",
			Description:       "Tests for basic Ethereum RPC methods without contract requirements",
			RequiresContracts: false,
			TestCases: []pkgTypes.TestCase{
				NewBalanceTestCase(),
			},
		}, */
		{
			Name:              "Contract Interaction Tests",
			Description:       "",
			RequiresContracts: true,
			TestCases: []pkgTypes.TestCase{
				// NewCodeAtTestCase(),
				// NewCallTestCase(),
				NewEstimateGasTestCase(),
				NewSendTransactionTestCase(),
			},
		},
	}
}

func GetTestSuiteByName(name string) *pkgTypes.TestSuite {
	for _, suite := range GetAllTestSuites() {
		if suite.Name == name {
			return &suite
		}
	}
	return nil
}

func GetTestCaseByName(name string) pkgTypes.TestCase {
	testCaseMap := map[string]pkgTypes.TestCase{
		"eth_getBalance":         NewBalanceTestCase(),
		"eth_getCode":            NewCodeAtTestCase(),
		"eth_call":               NewCallTestCase(),
		"eth_estimateGas":        NewEstimateGasTestCase(),
		"eth_sendRawTransaction": NewSendTransactionTestCase(),
	}

	return testCaseMap[name]
}
