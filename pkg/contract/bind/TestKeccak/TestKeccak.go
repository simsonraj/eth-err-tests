package TestKeccak

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"

	"github.com/eth-error-tests/pkg/contract/generated"
)

var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

var TestKeccakMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"iterations\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"hash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"result\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"iterations\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"hashAndStore\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50610276806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c806353d4d0981461003b578063ba0d647114610060575b600080fd5b61004e610049366004610133565b610075565b60405190815260200160405180910390f35b61007361006e366004610133565b6100f3565b005b80516020820120600090815b61008c60018661021e565b63ffffffff168163ffffffff1610156100eb57604080516020810184905201604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe081840301815291905280516020909101209150600101610081565b509392505050565b6100fd8282610075565b6000555050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000806040838503121561014657600080fd5b823563ffffffff8116811461015a57600080fd5b9150602083013567ffffffffffffffff8082111561017757600080fd5b818501915085601f83011261018b57600080fd5b81358181111561019d5761019d610104565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f011681019083821181831017156101e3576101e3610104565b816040528281528860208487010111156101fc57600080fd5b8260208601602083013760006020848301015280955050505050509250929050565b63ffffffff828116828216039080821115610262577f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b509291505056fea164736f6c6343000818000a",
}

var TestKeccakABI = TestKeccakMetaData.ABI

var TestKeccakBin = TestKeccakMetaData.Bin

func DeployTestKeccak(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *generated.Transaction, *TestKeccak, error) {
	parsed, err := TestKeccakMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}
	if generated.IsZKSync(backend) {
		address, ethTx, contractBind, _ := generated.DeployContract(auth, parsed, common.FromHex(TestKeccakZKBin), backend)
		contractReturn := &TestKeccak{address: address, abi: *parsed, TestKeccakCaller: TestKeccakCaller{contract: contractBind}, TestKeccakTransactor: TestKeccakTransactor{contract: contractBind}, TestKeccakFilterer: TestKeccakFilterer{contract: contractBind}}
		return address, ethTx, contractReturn, err
	}
	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TestKeccakBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, &generated.Transaction{Transaction: tx, HashZks: tx.Hash()}, &TestKeccak{address: address, abi: *parsed, TestKeccakCaller: TestKeccakCaller{contract: contract}, TestKeccakTransactor: TestKeccakTransactor{contract: contract}, TestKeccakFilterer: TestKeccakFilterer{contract: contract}}, nil
}

type TestKeccak struct {
	address common.Address
	abi     abi.ABI
	TestKeccakCaller
	TestKeccakTransactor
	TestKeccakFilterer
}

type TestKeccakCaller struct {
	contract *bind.BoundContract
}

type TestKeccakTransactor struct {
	contract *bind.BoundContract
}

type TestKeccakFilterer struct {
	contract *bind.BoundContract
}

type TestKeccakSession struct {
	Contract     *TestKeccak
	CallOpts     bind.CallOpts
	TransactOpts bind.TransactOpts
}

type TestKeccakCallerSession struct {
	Contract *TestKeccakCaller
	CallOpts bind.CallOpts
}

type TestKeccakTransactorSession struct {
	Contract     *TestKeccakTransactor
	TransactOpts bind.TransactOpts
}

type TestKeccakRaw struct {
	Contract *TestKeccak
}

type TestKeccakCallerRaw struct {
	Contract *TestKeccakCaller
}

type TestKeccakTransactorRaw struct {
	Contract *TestKeccakTransactor
}

func NewTestKeccak(address common.Address, backend bind.ContractBackend) (*TestKeccak, error) {
	abi, err := abi.JSON(strings.NewReader(TestKeccakABI))
	if err != nil {
		return nil, err
	}
	contract, err := bindTestKeccak(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestKeccak{address: address, abi: abi, TestKeccakCaller: TestKeccakCaller{contract: contract}, TestKeccakTransactor: TestKeccakTransactor{contract: contract}, TestKeccakFilterer: TestKeccakFilterer{contract: contract}}, nil
}

func NewTestKeccakCaller(address common.Address, caller bind.ContractCaller) (*TestKeccakCaller, error) {
	contract, err := bindTestKeccak(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestKeccakCaller{contract: contract}, nil
}

func NewTestKeccakTransactor(address common.Address, transactor bind.ContractTransactor) (*TestKeccakTransactor, error) {
	contract, err := bindTestKeccak(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestKeccakTransactor{contract: contract}, nil
}

func NewTestKeccakFilterer(address common.Address, filterer bind.ContractFilterer) (*TestKeccakFilterer, error) {
	contract, err := bindTestKeccak(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestKeccakFilterer{contract: contract}, nil
}

func bindTestKeccak(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestKeccakMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

func (_TestKeccak *TestKeccakRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestKeccak.Contract.TestKeccakCaller.contract.Call(opts, result, method, params...)
}

func (_TestKeccak *TestKeccakRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestKeccak.Contract.TestKeccakTransactor.contract.Transfer(opts)
}

func (_TestKeccak *TestKeccakRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestKeccak.Contract.TestKeccakTransactor.contract.Transact(opts, method, params...)
}

func (_TestKeccak *TestKeccakCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestKeccak.Contract.contract.Call(opts, result, method, params...)
}

func (_TestKeccak *TestKeccakTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestKeccak.Contract.contract.Transfer(opts)
}

func (_TestKeccak *TestKeccakTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestKeccak.Contract.contract.Transact(opts, method, params...)
}

func (_TestKeccak *TestKeccakCaller) Hash(opts *bind.CallOpts, iterations uint32, data []byte) ([32]byte, error) {
	var out []interface{}
	err := _TestKeccak.contract.Call(opts, &out, "hash", iterations, data)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

func (_TestKeccak *TestKeccakSession) Hash(iterations uint32, data []byte) ([32]byte, error) {
	return _TestKeccak.Contract.Hash(&_TestKeccak.CallOpts, iterations, data)
}

func (_TestKeccak *TestKeccakCallerSession) Hash(iterations uint32, data []byte) ([32]byte, error) {
	return _TestKeccak.Contract.Hash(&_TestKeccak.CallOpts, iterations, data)
}

func (_TestKeccak *TestKeccakTransactor) HashAndStore(opts *bind.TransactOpts, iterations uint32, data []byte) (*types.Transaction, error) {
	return _TestKeccak.contract.Transact(opts, "hashAndStore", iterations, data)
}

func (_TestKeccak *TestKeccakSession) HashAndStore(iterations uint32, data []byte) (*types.Transaction, error) {
	return _TestKeccak.Contract.HashAndStore(&_TestKeccak.TransactOpts, iterations, data)
}

func (_TestKeccak *TestKeccakTransactorSession) HashAndStore(iterations uint32, data []byte) (*types.Transaction, error) {
	return _TestKeccak.Contract.HashAndStore(&_TestKeccak.TransactOpts, iterations, data)
}

func (_TestKeccak *TestKeccak) Address() common.Address {
	return _TestKeccak.address
}

type TestKeccakInterface interface {
	Hash(opts *bind.CallOpts, iterations uint32, data []byte) ([32]byte, error)

	HashAndStore(opts *bind.TransactOpts, iterations uint32, data []byte) (*types.Transaction, error)

	Address() common.Address
}

var TestKeccakZKBin = ("0x0001000000000002000000000301034f00000000000303550000008001000039000000400010043f00000001002001900000001e0000c13d000000000103001900000060011002700000003501100197000000040010008c0000002e0000413d000000000203043b000000e002200270000000370020009c000000260000613d000000380020009c0000002e0000c13d0000000002000416000000000002004b0000002e0000c13d00d100300000040f00d1007d0000040f000000400200043d0000000000120435000000350020009c0000003502008041000000400120021000000039011001c7000000d20001042e0000000001000416000000000001004b0000002e0000c13d0000002001000039000001000010044300000120000004430000003601000041000000d20001042e0000000002000416000000000002004b0000002e0000c13d00d100300000040f00d1007d0000040f000000000010041b0000000001000019000000d20001042e0000000001000019000000d30001043000000000030100190000003a0030009c000000750000213d000000430030008c000000750000a13d00000000050003670000000401500370000000000101043b000000350010009c000000750000213d0000002402500370000000000702043b0000003b0070009c000000750000213d0000002302700039000000000032004b000000750000813d0000000408700039000000000285034f000000000402043b0000003c0040009c000000770000813d0000001f024000390000003f022001970000003f022000390000003f06200197000000400200043d0000000006620019000000000026004b000000000a000039000000010a0040390000003b0060009c000000770000213d0000000100a00190000000770000c13d000000400060043f000000000642043600000000074700190000002407700039000000000037004b000000750000213d0000002003800039000000000535034f0000003f074001980000001f0840018f0000000003760019000000650000613d000000000905034f000000000a060019000000009b09043c000000000aba043600000000003a004b000000610000c13d000000000008004b000000720000613d000000000575034f0000000307800210000000000803043300000000087801cf000000000878022f000000000505043b0000010007700089000000000575022f00000000057501cf000000000585019f000000000053043500000000034600190000000000030435000000000001042d0000000001000019000000d3000104300000003d01000041000000000010043f0000004101000039000000040010043f0000003e01000041000000d3000104300002000000000002000200000001001d0000002001200039000000350010009c000000350100804100000040011002100000000002020433000000350020009c00000035020080410000006002200210000000000112019f0000000002000414000000350020009c0000003502008041000000c002200210000000000112019f00000040011001c7000080100200003900d100cc0000040f0000000100200190000000be0000613d00000002020000290000003502200197000000010220008a000000350020009c000000c60000213d000000000101043b000000000002004b000000bd0000613d0000000003000019000100000002001d000200000003001d000000400200043d000000200300003900000000033204360000000000130435000000410020009c000000c00000813d0000004001200039000000400010043f000000350030009c000000350300804100000040013002100000000002020433000000350020009c00000035020080410000006002200210000000000112019f0000000002000414000000350020009c0000003502008041000000c002200210000000000112019f00000040011001c7000080100200003900d100cc0000040f0000000100200190000000be0000613d000000000101043b000000020200002900000001022000390000003503200197000000010030006c0000009c0000413d000000000001042d0000000001000019000000d3000104300000003d01000041000000000010043f0000004101000039000000040010043f0000003e01000041000000d3000104300000003d01000041000000000010043f0000001101000039000000040010043f0000003e01000041000000d300010430000000cf002104230000000102000039000000000001042d0000000002000019000000000001042d000000d100000432000000d20001042e000000d30001043000000000000000000000000000000000000000000000000000000000ffffffff000000020000000000000000000000000000004000000100000000000000000000000000000000000000000000000000000000000000000000000000ba0d64710000000000000000000000000000000000000000000000000000000053d4d09800000000000000000000000000000000000000200000000000000000000000007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000000000000000000000000000000000000000000000ffffffffffffffff00000000000000000000000000000000000000000000000100000000000000004e487b71000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000024000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe00200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffffffffffffc00000000000000000000000000000000000000000000000000000000000000000")
