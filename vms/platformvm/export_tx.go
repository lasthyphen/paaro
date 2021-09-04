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
	safemath "github.com/djt-labs/paaro/utils/math"
	"github.com/djt-labs/paaro/vms/components/djtx"
	"github.com/djt-labs/paaro/vms/secp256k1fx"
)

var (
	errNoExportOutputs  = errors.New("no export outputs")
	errOutputsNotSorted = errors.New("outputs not sorted")
	errOverflowExport   = errors.New("overflow when computing export amount + txFee")
	errWrongChainID     = errors.New("tx has wrong chain ID")

	_ UnsignedAtomicTx = &UnsignedExportTx{}
)

// UnsignedExportTx is an unsigned ExportTx
type UnsignedExportTx struct {
	BaseTx `serialize:"true"`

	// Which chain to send the funds to
	DestinationChain ids.ID `serialize:"true" json:"destinationChain"`

	// Outputs that are exported to the chain
	ExportedOutputs []*djtx.TransferableOutput `serialize:"true" json:"exportedOutputs"`
}

// InputUTXOs returns an empty set
func (tx *UnsignedExportTx) InputUTXOs() ids.Set { return ids.Set{} }

// Verify this transaction is well-formed
func (tx *UnsignedExportTx) Verify(
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
	case tx.DestinationChain != dvmID:
		// TODO: remove this check if we allow for P->C swaps
		return errWrongChainID
	case len(tx.ExportedOutputs) == 0:
		return errNoExportOutputs
	}

	if err := tx.BaseTx.Verify(ctx, c); err != nil {
		return err
	}

	for _, out := range tx.ExportedOutputs {
		if err := out.Verify(); err != nil {
			return fmt.Errorf("output failed verification: %w", err)
		}
		if _, ok := out.Output().(*StakeableLockOut); ok {
			return errWrongLocktime
		}
	}
	if !djtx.IsSortedTransferableOutputs(tx.ExportedOutputs, Codec) {
		return errOutputsNotSorted
	}

	tx.syntacticallyVerified = true
	return nil
}

// SemanticVerify this transaction is valid.
func (tx *UnsignedExportTx) SemanticVerify(
	vm *VM,
	parentState MutableState,
	stx *Tx,
) (VersionedState, TxError) {
	if err := tx.Verify(vm.ctx.XChainID, vm.ctx, vm.codec, vm.TxFee, vm.ctx.DJTXAssetID); err != nil {
		return nil, permError{err}
	}

	outs := make([]*djtx.TransferableOutput, len(tx.Outs)+len(tx.ExportedOutputs))
	copy(outs, tx.Outs)
	copy(outs[len(tx.Outs):], tx.ExportedOutputs)

	// Verify the flowcheck
	if err := vm.semanticVerifySpend(parentState, tx, tx.Ins, outs, stx.Creds, vm.TxFee, vm.ctx.DJTXAssetID); err != nil {
		switch err.(type) {
		case permError:
			return nil, permError{
				fmt.Errorf("failed semanticVerifySpend: %w", err),
			}
		default:
			return nil, tempError{
				fmt.Errorf("failed semanticVerifySpend: %w", err),
			}
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

// Accept this transaction.
func (tx *UnsignedExportTx) Accept(ctx *snow.Context, batch database.Batch) error {
	txID := tx.ID()

	elems := make([]*atomic.Element, len(tx.ExportedOutputs))
	for i, out := range tx.ExportedOutputs {
		utxo := &djtx.UTXO{
			UTXOID: djtx.UTXOID{
				TxID:        txID,
				OutputIndex: uint32(len(tx.Outs) + i),
			},
			Asset: djtx.Asset{ID: out.AssetID()},
			Out:   out.Out,
		}

		utxoBytes, err := Codec.Marshal(codecVersion, utxo)
		if err != nil {
			return fmt.Errorf("failed to marshal UTXO: %w", err)
		}
		utxoID := utxo.InputID()
		elem := &atomic.Element{
			Key:   utxoID[:],
			Value: utxoBytes,
		}
		if out, ok := utxo.Out.(djtx.Addressable); ok {
			elem.Traits = out.Addresses()
		}

		elems[i] = elem
	}

	return ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{tx.DestinationChain: {PutRequests: elems}}, batch)
}

// Create a new transaction
func (vm *VM) newExportTx(
	amount uint64, // Amount of tokens to export
	chainID ids.ID, // Chain to send the UTXOs to
	to ids.ShortID, // Address of chain recipient
	keys []*crypto.PrivateKeySECP256K1R, // Pay the fee and provide the tokens
	changeAddr ids.ShortID, // Address to send change to, if there is any
) (*Tx, error) {
	if vm.ctx.XChainID != chainID {
		return nil, errWrongChainID
	}

	toBurn, err := safemath.Add64(amount, vm.TxFee)
	if err != nil {
		return nil, errOverflowExport
	}
	ins, outs, _, signers, err := vm.stake(keys, 0, toBurn, changeAddr)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate tx inputs/outputs: %w", err)
	}

	// Create the transaction
	utx := &UnsignedExportTx{
		BaseTx: BaseTx{BaseTx: djtx.BaseTx{
			NetworkID:    vm.ctx.NetworkID,
			BlockchainID: vm.ctx.ChainID,
			Ins:          ins,
			Outs:         outs, // Non-exported outputs
		}},
		DestinationChain: chainID,
		ExportedOutputs: []*djtx.TransferableOutput{{ // Exported to X-Chain
			Asset: djtx.Asset{ID: vm.ctx.DJTXAssetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: amount,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{to},
				},
			},
		}},
	}
	tx := &Tx{UnsignedTx: utx}
	if err := tx.Sign(vm.codec, signers); err != nil {
		return nil, err
	}
	return tx, utx.Verify(vm.ctx.XChainID, vm.ctx, vm.codec, vm.TxFee, vm.ctx.DJTXAssetID)
}
