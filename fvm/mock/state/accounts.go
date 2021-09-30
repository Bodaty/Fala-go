// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	atree "github.com/onflow/atree"
	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// Accounts is an autogenerated mock type for the Accounts type
type Accounts struct {
	mock.Mock
}

// AllocateStorageIndex provides a mock function with given fields: address
func (_m *Accounts) AllocateStorageIndex(address flow.Address) (atree.StorageIndex, error) {
	ret := _m.Called(address)

	var r0 atree.StorageIndex
	if rf, ok := ret.Get(0).(func(flow.Address) atree.StorageIndex); ok {
		r0 = rf(address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(atree.StorageIndex)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AppendPublicKey provides a mock function with given fields: address, key
func (_m *Accounts) AppendPublicKey(address flow.Address, key flow.AccountPublicKey) error {
	ret := _m.Called(address, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Address, flow.AccountPublicKey) error); ok {
		r0 = rf(address, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CheckAccountNotFrozen provides a mock function with given fields: address
func (_m *Accounts) CheckAccountNotFrozen(address flow.Address) error {
	ret := _m.Called(address)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Address) error); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: publicKeys, newAddress
func (_m *Accounts) Create(publicKeys []flow.AccountPublicKey, newAddress flow.Address) error {
	ret := _m.Called(publicKeys, newAddress)

	var r0 error
	if rf, ok := ret.Get(0).(func([]flow.AccountPublicKey, flow.Address) error); ok {
		r0 = rf(publicKeys, newAddress)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteContract provides a mock function with given fields: contractName, address
func (_m *Accounts) DeleteContract(contractName string, address flow.Address) error {
	ret := _m.Called(contractName, address)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, flow.Address) error); ok {
		r0 = rf(contractName, address)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: address
func (_m *Accounts) Exists(address flow.Address) (bool, error) {
	ret := _m.Called(address)

	var r0 bool
	if rf, ok := ret.Get(0).(func(flow.Address) bool); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContract provides a mock function with given fields: contractName, address
func (_m *Accounts) GetContract(contractName string, address flow.Address) ([]byte, error) {
	ret := _m.Called(contractName, address)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, flow.Address) []byte); ok {
		r0 = rf(contractName, address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, flow.Address) error); ok {
		r1 = rf(contractName, address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContractNames provides a mock function with given fields: address
func (_m *Accounts) GetContractNames(address flow.Address) ([]string, error) {
	ret := _m.Called(address)

	var r0 []string
	if rf, ok := ret.Get(0).(func(flow.Address) []string); ok {
		r0 = rf(address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPublicKey provides a mock function with given fields: address, keyIndex
func (_m *Accounts) GetPublicKey(address flow.Address, keyIndex uint64) (flow.AccountPublicKey, error) {
	ret := _m.Called(address, keyIndex)

	var r0 flow.AccountPublicKey
	if rf, ok := ret.Get(0).(func(flow.Address, uint64) flow.AccountPublicKey); ok {
		r0 = rf(address, keyIndex)
	} else {
		r0 = ret.Get(0).(flow.AccountPublicKey)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address, uint64) error); ok {
		r1 = rf(address, keyIndex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPublicKeyCount provides a mock function with given fields: address
func (_m *Accounts) GetPublicKeyCount(address flow.Address) (uint64, error) {
	ret := _m.Called(address)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(flow.Address) uint64); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStorageUsed provides a mock function with given fields: address
func (_m *Accounts) GetStorageUsed(address flow.Address) (uint64, error) {
	ret := _m.Called(address)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(flow.Address) uint64); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetValue provides a mock function with given fields: address, key
func (_m *Accounts) GetValue(address flow.Address, key string) ([]byte, error) {
	ret := _m.Called(address, key)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(flow.Address, string) []byte); ok {
		r0 = rf(address, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address, string) error); ok {
		r1 = rf(address, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetAccountFrozen provides a mock function with given fields: address, frozen
func (_m *Accounts) SetAccountFrozen(address flow.Address, frozen bool) error {
	ret := _m.Called(address, frozen)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Address, bool) error); ok {
		r0 = rf(address, frozen)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetContract provides a mock function with given fields: contractName, address, contract
func (_m *Accounts) SetContract(contractName string, address flow.Address, contract []byte) error {
	ret := _m.Called(contractName, address, contract)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, flow.Address, []byte) error); ok {
		r0 = rf(contractName, address, contract)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPublicKey provides a mock function with given fields: address, keyIndex, publicKey
func (_m *Accounts) SetPublicKey(address flow.Address, keyIndex uint64, publicKey flow.AccountPublicKey) ([]byte, error) {
	ret := _m.Called(address, keyIndex, publicKey)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(flow.Address, uint64, flow.AccountPublicKey) []byte); ok {
		r0 = rf(address, keyIndex, publicKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Address, uint64, flow.AccountPublicKey) error); ok {
		r1 = rf(address, keyIndex, publicKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetValue provides a mock function with given fields: address, key, value
func (_m *Accounts) SetValue(address flow.Address, key string, value []byte) error {
	ret := _m.Called(address, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Address, string, []byte) error); ok {
		r0 = rf(address, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
