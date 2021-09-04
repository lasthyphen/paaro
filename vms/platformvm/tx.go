// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"fmt"

	"github.com/lasthyphen/paaro/codec"
	"github.com/lasthyphen/paaro/database"
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow"
	"github.com/lasthyphen/paaro/utils/crypto"
	"github.com/lasthyphen/paaro/utils/hashing"
	"github.com/lasthyphen/paaro/vms/components/verify"
	"github.com/lasthyphen/paaro/vms/secp256k1fx"
)

// UnsignedTx is an unsigned transaction
type UnsignedTx interface {
	Initialize(unsignedBytes, signedBytes []byte)
	ID() ids.ID
	UnsignedBytes() []byte
	Bytes() []byte
}

// UnsignedDecisionTx is an unsigned operation that can be immediately decided
type UnsignedDecisionTx interface {
	UnsignedTx

	// Attempts to verify this transaction with the provided state.
	SemanticVerify(vm *VM, vs VersionedState, stx *Tx) (
		onAcceptFunc func() error,
		err TxError,
	)
}

// UnsignedProposalTx is an unsigned operation that can be proposed
type UnsignedProposalTx interface {
	UnsignedTx

	// Attempts to verify this transaction with the provided state.
	SemanticVerify(vm *VM, state MutableState, stx *Tx) (
		onCommitState VersionedState,
		onAbortState VersionedState,
		onCommitFunc func() error,
		onAbortFunc func() error,
		err TxError,
	)
	InitiallyPrefersCommit(vm *VM) bool
}

// UnsignedAtomicTx is an unsigned operation that can be atomically accepted
type UnsignedAtomicTx interface {
	UnsignedTx

	// UTXOs this tx consumes
	InputUTXOs() ids.Set
	// Attempts to verify this transaction with the provided state.
	SemanticVerify(vm *VM, parentState MutableState, stx *Tx) (VersionedState, TxError)

	// Accept this transaction with the additionally provided state transitions.
	Accept(ctx *snow.Context, batch database.Batch) error
}

// Tx is a signed transaction
type Tx struct {
	// The body of this transaction
	UnsignedTx `serialize:"true" json:"unsignedTx"`

	// The credentials of this transaction
	Creds []verify.Verifiable `serialize:"true" json:"credentials"`
}

// Sign this transaction with the provided signers
func (tx *Tx) Sign(c codec.Manager, signers [][]*crypto.PrivateKeySECP256K1R) error {
	unsignedBytes, err := c.Marshal(codecVersion, &tx.UnsignedTx)
	if err != nil {
		return fmt.Errorf("couldn't marshal UnsignedTx: %w", err)
	}

	// Attach credentials
	hash := hashing.ComputeHash256(unsignedBytes)
	for _, keys := range signers {
		cred := &secp256k1fx.Credential{
			Sigs: make([][crypto.SECP256K1RSigLen]byte, len(keys)),
		}
		for i, key := range keys {
			sig, err := key.SignHash(hash) // Sign hash
			if err != nil {
				return fmt.Errorf("problem generating credential: %w", err)
			}
			copy(cred.Sigs[i][:], sig)
		}
		tx.Creds = append(tx.Creds, cred) // Attach credential
	}

	signedBytes, err := c.Marshal(codecVersion, tx)
	if err != nil {
		return fmt.Errorf("couldn't marshal ProposalTx: %w", err)
	}
	tx.Initialize(unsignedBytes, signedBytes)
	return nil
}
