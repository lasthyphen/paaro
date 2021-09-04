// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package secp256k1fx

import (
	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/utils/logging"
	"github.com/djt-labs/paaro/utils/timer"
)

// VM that this Fx must be run by
type VM interface {
	CodecRegistry() codec.Registry
	Clock() *timer.Clock
	Logger() logging.Logger
}

var _ VM = &TestVM{}

// TestVM is a minimal implementation of a VM
type TestVM struct {
	CLK   timer.Clock
	Codec codec.Registry
	Log   logging.Logger
}

func (vm *TestVM) Clock() *timer.Clock           { return &vm.CLK }
func (vm *TestVM) CodecRegistry() codec.Registry { return vm.Codec }
func (vm *TestVM) Logger() logging.Logger        { return vm.Log }
