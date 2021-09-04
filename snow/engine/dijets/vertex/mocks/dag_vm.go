// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	ids "github.com/lasthyphen/paaro/ids"
	common "github.com/lasthyphen/paaro/snow/engine/common"

	manager "github.com/lasthyphen/paaro/database/manager"

	mock "github.com/stretchr/testify/mock"

	snow "github.com/lasthyphen/paaro/snow"

	snowstorm "github.com/lasthyphen/paaro/snow/consensus/snowstorm"
)

// DAGVM is an autogenerated mock type for the DAGVM type
type DAGVM struct {
	mock.Mock
}

// Bootstrapped provides a mock function with given fields:
func (_m *DAGVM) Bootstrapped() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Bootstrapping provides a mock function with given fields:
func (_m *DAGVM) Bootstrapping() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateHandlers provides a mock function with given fields:
func (_m *DAGVM) CreateHandlers() (map[string]*common.HTTPHandler, error) {
	ret := _m.Called()

	var r0 map[string]*common.HTTPHandler
	if rf, ok := ret.Get(0).(func() map[string]*common.HTTPHandler); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*common.HTTPHandler)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTx provides a mock function with given fields: _a0
func (_m *DAGVM) GetTx(_a0 ids.ID) (snowstorm.Tx, error) {
	ret := _m.Called(_a0)

	var r0 snowstorm.Tx
	if rf, ok := ret.Get(0).(func(ids.ID) snowstorm.Tx); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(snowstorm.Tx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ids.ID) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HealthCheck provides a mock function with given fields:
func (_m *DAGVM) HealthCheck() (interface{}, error) {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Initialize provides a mock function with given fields: ctx, dbManager, genesisBytes, upgradeBytes, configBytes, toEngine, fxs
func (_m *DAGVM) Initialize(ctx *snow.Context, dbManager manager.Manager, genesisBytes []byte, upgradeBytes []byte, configBytes []byte, toEngine chan<- common.Message, fxs []*common.Fx) error {
	ret := _m.Called(ctx, dbManager, genesisBytes, upgradeBytes, configBytes, toEngine, fxs)

	var r0 error
	if rf, ok := ret.Get(0).(func(*snow.Context, manager.Manager, []byte, []byte, []byte, chan<- common.Message, []*common.Fx) error); ok {
		r0 = rf(ctx, dbManager, genesisBytes, upgradeBytes, configBytes, toEngine, fxs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ParseTx provides a mock function with given fields: tx
func (_m *DAGVM) ParseTx(tx []byte) (snowstorm.Tx, error) {
	ret := _m.Called(tx)

	var r0 snowstorm.Tx
	if rf, ok := ret.Get(0).(func([]byte) snowstorm.Tx); ok {
		r0 = rf(tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(snowstorm.Tx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PendingTxs provides a mock function with given fields:
func (_m *DAGVM) PendingTxs() []snowstorm.Tx {
	ret := _m.Called()

	var r0 []snowstorm.Tx
	if rf, ok := ret.Get(0).(func() []snowstorm.Tx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]snowstorm.Tx)
		}
	}

	return r0
}

// Shutdown provides a mock function with given fields:
func (_m *DAGVM) Shutdown() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
