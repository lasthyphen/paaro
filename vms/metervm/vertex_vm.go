// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"fmt"

	"github.com/djt-labs/paaro/database/manager"
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/snow/consensus/snowstorm"
	"github.com/djt-labs/paaro/snow/engine/dijets/vertex"
	"github.com/djt-labs/paaro/snow/engine/common"
	"github.com/djt-labs/paaro/utils/timer"
)

var _ vertex.DAGVM = &vertexVM{}

func NewVertexVM(vm vertex.DAGVM) vertex.DAGVM {
	return &vertexVM{
		DAGVM: vm,
	}
}

type vertexVM struct {
	vertex.DAGVM
	vertexMetrics
	clock timer.Clock
}

func (vm *vertexVM) Initialize(
	ctx *snow.Context,
	db manager.Manager,
	genesisBytes,
	upgradeBytes,
	configBytes []byte,
	toEngine chan<- common.Message,
	fxs []*common.Fx,
) error {
	if err := vm.vertexMetrics.Initialize(fmt.Sprintf("metervm_%s", ctx.Namespace), ctx.Metrics); err != nil {
		return err
	}

	return vm.DAGVM.Initialize(ctx, db, genesisBytes, upgradeBytes, configBytes, toEngine, fxs)
}

func (vm *vertexVM) PendingTxs() []snowstorm.Tx {
	start := vm.clock.Time()
	txs := vm.DAGVM.PendingTxs()
	end := vm.clock.Time()
	vm.vertexMetrics.pending.Observe(float64(end.Sub(start)))
	return txs
}

func (vm *vertexVM) ParseTx(b []byte) (snowstorm.Tx, error) {
	start := vm.clock.Time()
	tx, err := vm.DAGVM.ParseTx(b)
	end := vm.clock.Time()
	vm.vertexMetrics.parse.Observe(float64(end.Sub(start)))
	return tx, err
}

func (vm *vertexVM) GetTx(txID ids.ID) (snowstorm.Tx, error) {
	start := vm.clock.Time()
	tx, err := vm.DAGVM.GetTx(txID)
	end := vm.clock.Time()
	vm.vertexMetrics.get.Observe(float64(end.Sub(start)))
	return tx, err
}
