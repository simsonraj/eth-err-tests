# Ethereum Clients Error Codes Test suite

Test suite for validating error code implementations across different Ethereum clients.

## Setup
The tests can be run against different Ethereum clients locally using docker or an Actual RPCs

1. Set your private key:
```bash
export PRIVATE_KEY=your_private_key_here
```

2. Run tests:
```bash
go run main.go -client geth-local > logs/geth-local.log
```

## Add New Clients

Edit `pkg/config/config.go`:

```go
myClientConfig = Config{
    Network:       "my-client",
    Url:           "http://localhost:8545",
    From:          "0xYourAddress",
    PrivateKey:    os.Getenv("PRIVATE_KEY"),
    ChainID:       1337,
    LocalNodeType: "myclient", // Will need to configure the docker setup as well to run this client
}
```

Add to `GetConfig()` function:
```go
case "myclient":
    return myClientConfig, nil
```

Add the new client docker setup in localnode/manager.go

## Generate Reports

Convert logs to CSV:
```bash
go run main.go -report logs/geth-local.log
```

Output: `geth-local.csv`
