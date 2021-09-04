// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"fmt"

	"github.com/lasthyphen/paaro/database/manager"
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow"
	"github.com/lasthyphen/paaro/snow/consensus/snowman"
	"github.com/lasthyphen/paaro/snow/engine/common"
	"github.com/lasthyphen/paaro/snow/engine/snowman/block"
	"github.com/lasthyphen/paaro/utils/timer"
)

var _ block.ChainVM = &blockVM{}

func NewBlockVM(vm block.ChainVM) block.ChainVM {
	return &blockVM{
		ChainVM: vm,
	}
}

type blockVM struct {
	block.ChainVM
	blockMetrics
	clock timer.Clock
}

func (vm *blockVM) Initialize(
	ctx *snow.Context,
	db manager.Manager,
	genesisBytes,
	upgradeBytes,
	configBytes []byte,
	toEngine chan<- common.Message,
	fxs []*common.Fx,
) error {
	if err := vm.blockMetrics.Initialize(fmt.Sprintf("metervm_%s", ctx.Namespace), ctx.Metrics); err != nil {
		return err
	}

	return vm.ChainVM.Initialize(ctx, db, genesisBytes, upgradeBytes, configBytes, toEngine, fxs)
}

func (vm *blockVM) BuildBlock() (snowman.Block, error) {
	start := vm.clock.Time()
	blk, err := vm.ChainVM.BuildBlock()
	end := vm.clock.Time()
	vm.blockMetrics.buildBlock.Observe(float64(end.Sub(start)))
	return blk, err
}

func (vm *blockVM) ParseBlock(b []byte) (snowman.Block, error) {
	start := vm.clock.Time()
	blk, err := vm.ChainVM.ParseBlock(b)
	end := vm.clock.Time()
	vm.blockMetrics.parseBlock.Observe(float64(end.Sub(start)))
	return blk, err
}

func (vm *blockVM) GetBlock(id ids.ID) (snowman.Block, error) {
	start := vm.clock.Time()
	blk, err := vm.ChainVM.GetBlock(id)
	end := vm.clock.Time()
	vm.blockMetrics.getBlock.Observe(float64(end.Sub(start)))
	return blk, err
}

func (vm *blockVM) SetPreference(id ids.ID) error {
	start := vm.clock.Time()
	err := vm.ChainVM.SetPreference(id)
	end := vm.clock.Time()
	vm.blockMetrics.setPreference.Observe(float64(end.Sub(start)))
	return err
}

func (vm *blockVM) LastAccepted() (ids.ID, error) {
	start := vm.clock.Time()
	lastAcceptedID, err := vm.ChainVM.LastAccepted()
	end := vm.clock.Time()
	vm.blockMetrics.lastAccepted.Observe(float64(end.Sub(start)))
	return lastAcceptedID, err
}
