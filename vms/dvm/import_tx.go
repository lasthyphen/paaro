// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dvm

import (
	"errors"

	"github.com/djt-labs/paaro/chains/atomic"
	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/database"
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/vms/components/djtx"
	"github.com/djt-labs/paaro/vms/components/verify"
)

var (
	errNoImportInputs = errors.New("no import inputs")

	_ UnsignedTx = &ImportTx{}
)

// ImportTx is a transaction that imports an asset from another blockchain.
type ImportTx struct {
	BaseTx `serialize:"true"`

	// Which chain to consume the funds from
	SourceChain ids.ID `serialize:"true" json:"sourceChain"`

	// The inputs to this transaction
	ImportedIns []*djtx.TransferableInput `serialize:"true" json:"importedInputs"`
}

func (t *ImportTx) Init(vm *VM) error {
	for _, in := range t.ImportedIns {
		fx, err := vm.getParsedFx(in.In)
		if err != nil {
			return err
		}
		in.FxID = fx.ID
	}
	return t.BaseTx.Init(vm)
}

// InputUTXOs track which UTXOs this transaction is consuming.
func (t *ImportTx) InputUTXOs() []*djtx.UTXOID {
	utxos := t.BaseTx.InputUTXOs()
	for _, in := range t.ImportedIns {
		in.Symbol = true
		utxos = append(utxos, &in.UTXOID)
	}
	return utxos
}

// ConsumedAssetIDs returns the IDs of the assets this transaction consumes
func (t *ImportTx) ConsumedAssetIDs() ids.Set {
	assets := t.BaseTx.AssetIDs()
	for _, in := range t.ImportedIns {
		assets.Add(in.AssetID())
	}
	return assets
}

// AssetIDs returns the IDs of the assets this transaction depends on
func (t *ImportTx) AssetIDs() ids.Set {
	assets := t.BaseTx.AssetIDs()
	for _, in := range t.ImportedIns {
		assets.Add(in.AssetID())
	}
	return assets
}

// NumCredentials returns the number of expected credentials
func (t *ImportTx) NumCredentials() int { return t.BaseTx.NumCredentials() + len(t.ImportedIns) }

// SyntacticVerify that this transaction is well-formed.
func (t *ImportTx) SyntacticVerify(
	ctx *snow.Context,
	c codec.Manager,
	txFeeAssetID ids.ID,
	txFee uint64,
	_ uint64,
	numFxs int,
) error {
	switch {
	case t == nil:
		return errNilTx
	case len(t.ImportedIns) == 0:
		return errNoImportInputs
	}

	if err := t.MetadataVerify(ctx); err != nil {
		return err
	}

	return djtx.VerifyTx(
		txFee,
		txFeeAssetID,
		[][]*djtx.TransferableInput{
			t.Ins,
			t.ImportedIns,
		},
		[][]*djtx.TransferableOutput{t.Outs},
		c,
	)
}

// SemanticVerify that this transaction is well-formed.
func (t *ImportTx) SemanticVerify(vm *VM, tx UnsignedTx, creds []verify.Verifiable) error {
	if err := t.BaseTx.SemanticVerify(vm, tx, creds); err != nil {
		return err
	}

	if !vm.bootstrapped {
		return nil
	}

	subnetID, err := vm.ctx.SNLookup.SubnetID(t.SourceChain)
	if err != nil {
		return err
	}
	if vm.ctx.SubnetID != subnetID || t.SourceChain == vm.ctx.ChainID {
		return errWrongBlockchainID
	}

	utxoIDs := make([][]byte, len(t.ImportedIns))
	for i, in := range t.ImportedIns {
		inputID := in.UTXOID.InputID()
		utxoIDs[i] = inputID[:]
	}
	allUTXOBytes, err := vm.ctx.SharedMemory.Get(t.SourceChain, utxoIDs)
	if err != nil {
		return err
	}

	offset := t.BaseTx.NumCredentials()
	for i, in := range t.ImportedIns {
		utxo := djtx.UTXO{}
		if _, err := vm.codec.Unmarshal(allUTXOBytes[i], &utxo); err != nil {
			return err
		}

		cred := creds[i+offset]

		if err := vm.verifyTransferOfUTXO(tx, in, cred, &utxo); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteWithSideEffects writes the batch with any additional side effects
func (t *ImportTx) ExecuteWithSideEffects(vm *VM, batch database.Batch) error {
	utxoIDs := make([][]byte, len(t.ImportedIns))
	for i, in := range t.ImportedIns {
		inputID := in.UTXOID.InputID()
		utxoIDs[i] = inputID[:]
	}
	return vm.ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{t.SourceChain: {RemoveRequests: utxoIDs}}, batch)
}
