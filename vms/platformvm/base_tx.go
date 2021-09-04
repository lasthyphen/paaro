package platformvm

import (
	"fmt"

	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/vms/components/djtx"
)

// BaseTx contains fields common to many transaction types. It should be
// embedded in transaction implementations.
type BaseTx struct {
	djtx.BaseTx `serialize:"true" json:"inputs"`

	// true iff this transaction has already passed syntactic verification
	syntacticallyVerified bool
}

// Verify returns nil iff this tx is well formed
func (tx *BaseTx) Verify(ctx *snow.Context, c codec.Manager) error {
	switch {
	case tx == nil:
		return errNilTx
	case tx.syntacticallyVerified: // already passed syntactic verification
		return nil
	}
	if err := tx.MetadataVerify(ctx); err != nil {
		return fmt.Errorf("metadata failed verification: %w", err)
	}
	for _, out := range tx.Outs {
		if err := out.Verify(); err != nil {
			return fmt.Errorf("output failed verification: %w", err)
		}
	}
	for _, in := range tx.Ins {
		if err := in.Verify(); err != nil {
			return fmt.Errorf("input failed verification: %w", err)
		}
	}
	switch {
	case !djtx.IsSortedTransferableOutputs(tx.Outs, c):
		return errOutputsNotSorted
	case !djtx.IsSortedAndUniqueTransferableInputs(tx.Ins):
		return errInputsNotSortedUnique
	default:
		return nil
	}
}
