package opCodes

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

var OpCodesMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"constant\":false,\"inputs\":[],\"name\":\"test\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"test_invalid\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"test_revert\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"test_stop\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5060405161001d9061005f565b604051809103906000f080158015610039573d6000803e3d6000fd5b50600080546001600160a01b0319166001600160a01b039290921691909117905561006c565b610113806103b783390190565b61033c8061007b6000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c806355313dea146100515780636d3d141614610057578063b9d1e5aa1461005f578063f8a8fd6d14610067575b600080fd5b6100555b005b61005561004c565b61005561006f565b610055610055565bfe5b600581101561008257600101610071565b5060065b600581111561009757600101610086565b5060015b60058112156100ac5760010161009b565b5060065b60058113156100c1576001016100b0565b5060405160208101602060048337505060405160208101602060048339505060405160208101602060048360003c50503660005b8181101561010957600281526001016100f5565b5060009050602060403e6010608060106040610123612710fa506020610123600af05060008054604080517f697353616d654164647265737328616464726573732c616464726573732900008152815190819003601e0181208082523360048301819052602483018190526064830190935273ffffffffffffffffffffffffffffffffffffffff909316936020908290604490829088611388f1505060405182815281600482015281602482015260648101604052602081604483600088611388f250506040518281528160048201528160248201526064810160405260208160448387611388f45050604080517f50cb9fe53daa9737b786ab3646f04d0150dc50ef4e75f59509d83667ad5adb2081529051624200429181900360200190a0604080517f50cb9fe53daa9737b786ab3646f04d0150dc50ef4e75f59509d83667ad5adb2080825291519081900360200190a1604080517f50cb9fe53daa9737b786ab3646f04d0150dc50ef4e75f59509d83667ad5adb20808252915133929181900360200190a2604080517f50cb9fe53daa9737b786ab3646f04d0150dc50ef4e75f59509d83667ad5adb2080825291518392339290919081900360200190a3604080517f50cb9fe53daa9737b786ab3646f04d0150dc50ef4e75f59509d83667ad5adb2080825291518392839233929081900360200190a46002fffea265627a7a72315820b7ed38a2fafc01b6a95a35dbc64af86ac719474cd38a8a01f6500118f2aa706464736f6c63430005100032608060405234801561001057600080fd5b5060f48061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063161e715014602d575b600080fd5b606560048036036040811015604157600080fd5b5073ffffffffffffffffffffffffffffffffffffffff813581169160200135166079565b604080519115158252519081900360200190f35b60008173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141560b55750600160b9565b5060005b9291505056fea265627a7a723158203d148be69b6bc901b2c877f52c9ce015dae5fb880f87cfc9dfc08846f9dc3a4c64736f6c63430005100032",
}

var OpCodesABI = OpCodesMetaData.ABI

var OpCodesBin = OpCodesMetaData.Bin

func DeployOpCodes(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *generated.Transaction, *OpCodes, error) {
	parsed, err := OpCodesMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}
	if generated.IsZKSync(backend) {
		address, ethTx, contractBind, _ := generated.DeployContract(auth, parsed, common.FromHex(OpCodesZKBin), backend)
		contractReturn := &OpCodes{address: address, abi: *parsed, OpCodesCaller: OpCodesCaller{contract: contractBind}, OpCodesTransactor: OpCodesTransactor{contract: contractBind}, OpCodesFilterer: OpCodesFilterer{contract: contractBind}}
		return address, ethTx, contractReturn, err
	}
	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(OpCodesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, &generated.Transaction{Transaction: tx, HashZks: tx.Hash()}, &OpCodes{address: address, abi: *parsed, OpCodesCaller: OpCodesCaller{contract: contract}, OpCodesTransactor: OpCodesTransactor{contract: contract}, OpCodesFilterer: OpCodesFilterer{contract: contract}}, nil
}

type OpCodes struct {
	address common.Address
	abi     abi.ABI
	OpCodesCaller
	OpCodesTransactor
	OpCodesFilterer
}

type OpCodesCaller struct {
	contract *bind.BoundContract
}

type OpCodesTransactor struct {
	contract *bind.BoundContract
}

type OpCodesFilterer struct {
	contract *bind.BoundContract
}

type OpCodesSession struct {
	Contract     *OpCodes
	CallOpts     bind.CallOpts
	TransactOpts bind.TransactOpts
}

type OpCodesCallerSession struct {
	Contract *OpCodesCaller
	CallOpts bind.CallOpts
}

type OpCodesTransactorSession struct {
	Contract     *OpCodesTransactor
	TransactOpts bind.TransactOpts
}

type OpCodesRaw struct {
	Contract *OpCodes
}

type OpCodesCallerRaw struct {
	Contract *OpCodesCaller
}

type OpCodesTransactorRaw struct {
	Contract *OpCodesTransactor
}

func NewOpCodes(address common.Address, backend bind.ContractBackend) (*OpCodes, error) {
	abi, err := abi.JSON(strings.NewReader(OpCodesABI))
	if err != nil {
		return nil, err
	}
	contract, err := bindOpCodes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OpCodes{address: address, abi: abi, OpCodesCaller: OpCodesCaller{contract: contract}, OpCodesTransactor: OpCodesTransactor{contract: contract}, OpCodesFilterer: OpCodesFilterer{contract: contract}}, nil
}

func NewOpCodesCaller(address common.Address, caller bind.ContractCaller) (*OpCodesCaller, error) {
	contract, err := bindOpCodes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OpCodesCaller{contract: contract}, nil
}

func NewOpCodesTransactor(address common.Address, transactor bind.ContractTransactor) (*OpCodesTransactor, error) {
	contract, err := bindOpCodes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OpCodesTransactor{contract: contract}, nil
}

func NewOpCodesFilterer(address common.Address, filterer bind.ContractFilterer) (*OpCodesFilterer, error) {
	contract, err := bindOpCodes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OpCodesFilterer{contract: contract}, nil
}

func bindOpCodes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OpCodesMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

func (_OpCodes *OpCodesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OpCodes.Contract.OpCodesCaller.contract.Call(opts, result, method, params...)
}

func (_OpCodes *OpCodesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpCodes.Contract.OpCodesTransactor.contract.Transfer(opts)
}

func (_OpCodes *OpCodesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OpCodes.Contract.OpCodesTransactor.contract.Transact(opts, method, params...)
}

func (_OpCodes *OpCodesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OpCodes.Contract.contract.Call(opts, result, method, params...)
}

func (_OpCodes *OpCodesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpCodes.Contract.contract.Transfer(opts)
}

func (_OpCodes *OpCodesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OpCodes.Contract.contract.Transact(opts, method, params...)
}

func (_OpCodes *OpCodesTransactor) Test(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpCodes.contract.Transact(opts, "test")
}

func (_OpCodes *OpCodesSession) Test() (*types.Transaction, error) {
	return _OpCodes.Contract.Test(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactorSession) Test() (*types.Transaction, error) {
	return _OpCodes.Contract.Test(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactor) TestInvalid(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpCodes.contract.Transact(opts, "test_invalid")
}

func (_OpCodes *OpCodesSession) TestInvalid() (*types.Transaction, error) {
	return _OpCodes.Contract.TestInvalid(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactorSession) TestInvalid() (*types.Transaction, error) {
	return _OpCodes.Contract.TestInvalid(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactor) TestRevert(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpCodes.contract.Transact(opts, "test_revert")
}

func (_OpCodes *OpCodesSession) TestRevert() (*types.Transaction, error) {
	return _OpCodes.Contract.TestRevert(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactorSession) TestRevert() (*types.Transaction, error) {
	return _OpCodes.Contract.TestRevert(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactor) TestStop(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OpCodes.contract.Transact(opts, "test_stop")
}

func (_OpCodes *OpCodesSession) TestStop() (*types.Transaction, error) {
	return _OpCodes.Contract.TestStop(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodesTransactorSession) TestStop() (*types.Transaction, error) {
	return _OpCodes.Contract.TestStop(&_OpCodes.TransactOpts)
}

func (_OpCodes *OpCodes) Address() common.Address {
	return _OpCodes.address
}

type OpCodesInterface interface {
	Test(opts *bind.TransactOpts) (*types.Transaction, error)

	TestInvalid(opts *bind.TransactOpts) (*types.Transaction, error)

	TestRevert(opts *bind.TransactOpts) (*types.Transaction, error)

	TestStop(opts *bind.TransactOpts) (*types.Transaction, error)

	Address() common.Address
}

var OpCodesZKBin = ("0x0000008003000039000000400030043f00000001002001900000001d0000c13d00000060021002700000000b02200197000000040020008c000000250000413d000000000301043b0000000d033001970000000e0030009c000000250000c13d000000440220008a000000110020009c000000250000213d0000000002000416000000000002004b000000250000c13d0000000402100370000000000202043b0000002401100370000000000101043b000000000121013f0000000f0010019800000000010000390000000101006039000000800010043f0000001001000041000000280001042e0000000001000416000000000001004b000000250000c13d0000002001000039000001000010044300000120000004430000000c01000041000000280001042e000000000100001900000029000104300000002700000432000000280001042e00000029000104300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffffff0000000200000000000000000000000000000040000001000000000000000000ffffffff00000000000000000000000000000000000000000000000000000000161e715000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000020000000800000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffbfd17246b8d6ec481c7ca697ca8cfd8c8c7886e2abb55b19718219a7649d74176c")
