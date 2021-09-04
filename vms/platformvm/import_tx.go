// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"errors"
	"fmt"

	"github.com/djt-labs/paaro/chains/atomic"
	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/database"
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/utils/crypto"
	"github.com/djt-labs/paaro/utils/math"
	"github.com/djt-labs/paaro/vms/components/djtx"
	"github.com/djt-labs/paaro/vms/secp256k1fx"
)

var (
	errAssetIDMismatch          = errors.New("asset IDs in the input don't match the utxo")
	errWrongNumberOfCredentials = errors.New("should have the same number of credentials as inputs")
	errNoImportInputs           = errors.New("tx has no imported inputs")
	errInputsNotSortedUnique    = errors.New("inputs not sorted and unique")

	_ UnsignedAtomicTx = &UnsignedImportTx{}
)

// UnsignedImportTx is an unsigned ImportTx
type UnsignedImportTx struct {
	BaseTx `serialize:"true"`

	// Which chain to consume the funds from
	SourceChain ids.ID `serialize:"true" json:"sourceChain"`

	// Inputs that consume UTXOs produced on the chain
	ImportedInputs []*djtx.TransferableInput `serialize:"true" json:"importedInputs"`
}

// InputUTXOs returns the UTXOIDs of the imported funds
func (tx *UnsignedImportTx) InputUTXOs() ids.Set {
	set := ids.NewSet(len(tx.ImportedInputs))
	for _, in := range tx.ImportedInputs {
		set.Add(in.InputID())
	}
	return set
}

// Verify this transaction is well-formed
func (tx *UnsignedImportTx) Verify(
	dvmID ids.ID,
	ctx *snow.Context,
	c codec.Manager,
	feeAmount uint64,
	feeAssetID ids.ID,
) error {
	switch {
	case tx == nil:
		return errNilTx
	case tx.syntacticallyVerified: // already passed syntactic verification
		return nil
	case tx.SourceChain != dvmID:
		// TODO: remove this check if we allow for P->C swaps
		return errWrongChainID
	case len(tx.ImportedInputs) == 0:
		return errNoImportInputs
	}

	if err := tx.BaseTx.Verify(ctx, c); err != nil {
		return err
	}

	for _, in := range tx.ImportedInputs {
		if err := in.Verify(); err != nil {
			return fmt.Errorf("input failed verification: %w", err)
		}
	}
	if !djtx.IsSortedAndUniqueTransferableInputs(tx.ImportedInputs) {
		return errInputsNotSortedUnique
	}

	tx.syntacticallyVerified = true
	return nil
}

// SemanticVerify this transaction is valid.
func (tx *UnsignedImportTx) SemanticVerify(
	vm *VM,
	parentState MutableState,
	stx *Tx,
) (VersionedState, TxError) {
	if err := tx.Verify(vm.ctx.XChainID, vm.ctx, vm.codec, vm.TxFee, vm.ctx.DJTXAssetID); err != nil {
		return nil, permError{err}
	}

	utxos := make([]*djtx.UTXO, len(tx.Ins)+len(tx.ImportedInputs))
	for index, input := range tx.Ins {
		utxo, err := parentState.GetUTXO(input.InputID())
		if err != nil {
			return nil, tempError{
				fmt.Errorf("failed to get UTXO %s: %w", &input.UTXOID, err),
			}
		}
		utxos[index] = utxo
	}

	if vm.bootstrapped {
		utxoIDs := make([][]byte, len(tx.ImportedInputs))
		for i, in := range tx.ImportedInputs {
			utxoID := in.UTXOID.InputID()
			utxoIDs[i] = utxoID[:]
		}
		allUTXOBytes, err := vm.ctx.SharedMemory.Get(tx.SourceChain, utxoIDs)
		if err != nil {
			return nil, tempError{
				fmt.Errorf("failed to get shared memory: %w", err),
			}
		}

		for i, utxoBytes := range allUTXOBytes {
			utxo := &djtx.UTXO{}
			if _, err := vm.codec.Unmarshal(utxoBytes, utxo); err != nil {
				return nil, tempError{
					fmt.Errorf("failed to unmarshal UTXO: %w", err),
				}
			}
			utxos[i+len(tx.Ins)] = utxo
		}

		ins := make([]*djtx.TransferableInput, len(tx.Ins)+len(tx.ImportedInputs))
		copy(ins, tx.Ins)
		copy(ins[len(tx.Ins):], tx.ImportedInputs)

		if err := vm.semanticVerifySpendUTXOs(tx, utxos, ins, tx.Outs, stx.Creds, vm.TxFee, vm.ctx.DJTXAssetID); err != nil {
			return nil, err
		}
	}

	// Set up the state if this tx is committed
	newState := newVersionedState(
		parentState,
		parentState.CurrentStakerChainState(),
		parentState.PendingStakerChainState(),
	)
	// Consume the UTXOS
	consumeInputs(newState, tx.Ins)
	// Produce the UTXOS
	txID := tx.ID()
	produceOutputs(newState, txID, vm.ctx.DJTXAssetID, tx.Outs)
	return newState, nil
}

// Accept this transaction and spend imported inputs
// We spend imported UTXOs here rather than in semanticVerify because
// we don't want to remove an imported UTXO in semanticVerify
// only to have the transaction not be Accepted. This would be inconsistent.
// Recall that imported UTXOs are not kept in a versionDB.
func (tx *UnsignedImportTx) Accept(ctx *snow.Context, batch database.Batch) error {
	utxoIDs := make([][]byte, len(tx.ImportedInputs))
	for i, in := range tx.ImportedInputs {
		utxoID := in.InputID()
		utxoIDs[i] = utxoID[:]
	}
	return ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{tx.SourceChain: {RemoveRequests: utxoIDs}}, batch)
}

// Create a new transaction
func (vm *VM) newImportTx(
	chainID ids.ID, // chain to import from
	to ids.ShortID, // Address of recipient
	keys []*crypto.PrivateKeySECP256K1R, // Keys to import the funds
	changeAddr ids.ShortID, // Address to send change to, if there is any
) (*Tx, error) {
	if vm.ctx.XChainID != chainID {
		return nil, errWrongChainID
	}

	kc := secp256k1fx.NewKeychain()
	for _, key := range keys {
		kc.Add(key)
	}

	atomicUTXOs, _, _, err := vm.GetAtomicUTXOs(chainID, kc.Addresses(), ids.ShortEmpty, ids.Empty, -1)
	if err != nil {
		return nil, fmt.Errorf("problem retrieving atomic UTXOs: %w", err)
	}

	importedInputs := []*djtx.TransferableInput{}
	signers := [][]*crypto.PrivateKeySECP256K1R{}

	importedAmount := uint64(0)
	now := vm.clock.Unix()
	for _, utxo := range atomicUTXOs {
		if utxo.AssetID() != vm.ctx.DJTXAssetID {
			continue
		}
		inputIntf, utxoSigners, err := kc.Spend(utxo.Out, now)
		if err != nil {
			continue
		}
		input, ok := inputIntf.(djtx.TransferableIn)
		if !ok {
			continue
		}
		importedAmount, err = math.Add64(importedAmount, input.Amount())
		if err != nil {
			return nil, err
		}
		importedInputs = append(importedInputs, &djtx.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  utxo.Asset,
			In:     input,
		})
		signers = append(signers, utxoSigners)
	}
	djtx.SortTransferableInputsWithSigners(importedInputs, signers)

	if importedAmount == 0 {
		return nil, errNoFunds // No imported UTXOs were spendable
	}

	ins := []*djtx.TransferableInput{}
	outs := []*djtx.TransferableOutput{}
	if importedAmount < vm.TxFee { // imported amount goes toward paying tx fee
		var baseSigners [][]*crypto.PrivateKeySECP256K1R
		ins, outs, _, baseSigners, err = vm.stake(keys, 0, vm.TxFee-importedAmount, changeAddr)
		if err != nil {
			return nil, fmt.Errorf("couldn't generate tx inputs/outputs: %w", err)
		}
		signers = append(baseSigners, signers...)
	} else if importedAmount > vm.TxFee {
		outs = append(outs, &djtx.TransferableOutput{
			Asset: djtx.Asset{ID: vm.ctx.DJTXAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: importedAmount - vm.TxFee,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{to},
				},
			},
		})
	}

	// Create the transaction
	utx := &UnsignedImportTx{
		BaseTx: BaseTx{BaseTx: djtx.BaseTx{
			NetworkID:    vm.ctx.NetworkID,
			BlockchainID: vm.ctx.ChainID,
			Outs:         outs,
			Ins:          ins,
		}},
		SourceChain:    chainID,
		ImportedInputs: importedInputs,
	}
	tx := &Tx{UnsignedTx: utx}
	if err := tx.Sign(vm.codec, signers); err != nil {
		return nil, err
	}
	return tx, utx.Verify(vm.ctx.XChainID, vm.ctx, vm.codec, vm.TxFee, vm.ctx.DJTXAssetID)
}
