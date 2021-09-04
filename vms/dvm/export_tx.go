// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dvm

import (
	"errors"

	"github.com/lasthyphen/paaro/chains/atomic"
	"github.com/lasthyphen/paaro/codec"
	"github.com/lasthyphen/paaro/database"
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow"
	"github.com/lasthyphen/paaro/utils/constants"
	"github.com/lasthyphen/paaro/vms/components/djtx"
	"github.com/lasthyphen/paaro/vms/components/verify"
)

var (
	errNoExportOutputs = errors.New("no export outputs")

	_ UnsignedTx = &ExportTx{}
)

// ExportTx is a transaction that exports an asset to another blockchain.
type ExportTx struct {
	BaseTx `serialize:"true"`

	// Which chain to send the funds to
	DestinationChain ids.ID `serialize:"true" json:"destinationChain"`

	// The outputs this transaction is sending to the other chain
	ExportedOuts []*djtx.TransferableOutput `serialize:"true" json:"exportedOutputs"`
}

func (t *ExportTx) Init(vm *VM) error {
	for _, out := range t.ExportedOuts {
		fx, err := vm.getParsedFx(out.Out)
		if err != nil {
			return err
		}
		out.FxID = fx.ID
		out.InitCtx(vm.ctx)
	}
	return t.BaseTx.Init(vm)
}

// SyntacticVerify that this transaction is well-formed.
func (t *ExportTx) SyntacticVerify(
	ctx *snow.Context,
	c codec.Manager,
	txFeeAssetID ids.ID,
	txFee uint64,
	_ uint64,
	_ int,
) error {
	switch {
	case t == nil:
		return errNilTx
	case len(t.ExportedOuts) == 0:
		return errNoExportOutputs
	}

	if err := t.MetadataVerify(ctx); err != nil {
		return err
	}

	return djtx.VerifyTx(
		txFee,
		txFeeAssetID,
		[][]*djtx.TransferableInput{t.Ins},
		[][]*djtx.TransferableOutput{
			t.Outs,
			t.ExportedOuts,
		},
		c,
	)
}

// SemanticVerify that this transaction is valid to be spent.
func (t *ExportTx) SemanticVerify(vm *VM, tx UnsignedTx, creds []verify.Verifiable) error {
	if vm.bootstrapped {
		subnetID, err := vm.ctx.SNLookup.SubnetID(t.DestinationChain)
		if err != nil {
			return err
		}
		if vm.ctx.SubnetID != subnetID || t.DestinationChain == vm.ctx.ChainID {
			return errWrongBlockchainID
		}
	}

	for _, out := range t.ExportedOuts {
		fxIndex, err := vm.getFx(out.Out)
		if err != nil {
			return err
		}
		assetID := out.AssetID()
		if assetID != vm.ctx.DJTXAssetID && t.DestinationChain == constants.PlatformChainID {
			return errWrongAssetID
		}
		if !vm.verifyFxUsage(fxIndex, assetID) {
			return errIncompatibleFx
		}
	}

	return t.BaseTx.SemanticVerify(vm, tx, creds)
}

// ExecuteWithSideEffects writes the batch with any additional side effects
func (t *ExportTx) ExecuteWithSideEffects(vm *VM, batch database.Batch) error {
	txID := t.ID()

	elems := make([]*atomic.Element, len(t.ExportedOuts))
	for i, out := range t.ExportedOuts {
		utxo := &djtx.UTXO{
			UTXOID: djtx.UTXOID{
				TxID:        txID,
				OutputIndex: uint32(len(t.Outs) + i),
			},
			Asset: djtx.Asset{ID: out.AssetID()},
			Out:   out.Out,
		}

		utxoBytes, err := vm.codec.Marshal(codecVersion, utxo)
		if err != nil {
			return err
		}

		inputID := utxo.InputID()
		elem := &atomic.Element{
			Key:   inputID[:],
			Value: utxoBytes,
		}
		if out, ok := utxo.Out.(djtx.Addressable); ok {
			elem.Traits = out.Addresses()
		}

		elems[i] = elem
	}

	return vm.ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{t.DestinationChain: {PutRequests: elems}}, batch)
}
