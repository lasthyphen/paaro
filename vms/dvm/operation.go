// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dvm

import (
	"bytes"
	"errors"
	"sort"

	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/utils"
	"github.com/djt-labs/paaro/utils/crypto"
	"github.com/djt-labs/paaro/vms/components/djtx"
	"github.com/djt-labs/paaro/vms/components/verify"
)

var (
	errNilOperation              = errors.New("nil operation is not valid")
	errNilFxOperation            = errors.New("nil fx operation is not valid")
	errNotSortedAndUniqueUTXOIDs = errors.New("utxo IDs not sorted and unique")
)

type Operation struct {
	djtx.Asset `serialize:"true"`
	UTXOIDs    []*djtx.UTXOID `serialize:"true" json:"inputIDs"`
	FxID       ids.ID         `serialize:"false" json:"fxID"`
	Op         FxOperation    `serialize:"true" json:"operation"`
}

// Verify implements the verify.Verifiable interface
func (op *Operation) Verify(c codec.Manager) error {
	switch {
	case op == nil:
		return errNilOperation
	case op.Op == nil:
		return errNilFxOperation
	case !djtx.IsSortedAndUniqueUTXOIDs(op.UTXOIDs):
		return errNotSortedAndUniqueUTXOIDs
	default:
		return verify.All(&op.Asset, op.Op)
	}
}

type innerSortOperation struct {
	ops   []*Operation
	codec codec.Manager
}

func (ops *innerSortOperation) Less(i, j int) bool {
	iOp := ops.ops[i]
	jOp := ops.ops[j]

	iBytes, err := ops.codec.Marshal(codecVersion, iOp)
	if err != nil {
		return false
	}
	jBytes, err := ops.codec.Marshal(codecVersion, jOp)
	if err != nil {
		return false
	}
	return bytes.Compare(iBytes, jBytes) == -1
}
func (ops *innerSortOperation) Len() int      { return len(ops.ops) }
func (ops *innerSortOperation) Swap(i, j int) { o := ops.ops; o[j], o[i] = o[i], o[j] }

func sortOperations(ops []*Operation, c codec.Manager) {
	sort.Sort(&innerSortOperation{ops: ops, codec: c})
}

func isSortedAndUniqueOperations(ops []*Operation, c codec.Manager) bool {
	return utils.IsSortedAndUnique(&innerSortOperation{ops: ops, codec: c})
}

type innerSortOperationsWithSigners struct {
	ops     []*Operation
	signers [][]*crypto.PrivateKeySECP256K1R
	codec   codec.Manager
}

func (ops *innerSortOperationsWithSigners) Less(i, j int) bool {
	iOp := ops.ops[i]
	jOp := ops.ops[j]

	iBytes, err := ops.codec.Marshal(codecVersion, iOp)
	if err != nil {
		return false
	}
	jBytes, err := ops.codec.Marshal(codecVersion, jOp)
	if err != nil {
		return false
	}
	return bytes.Compare(iBytes, jBytes) == -1
}
func (ops *innerSortOperationsWithSigners) Len() int { return len(ops.ops) }
func (ops *innerSortOperationsWithSigners) Swap(i, j int) {
	ops.ops[j], ops.ops[i] = ops.ops[i], ops.ops[j]
	ops.signers[j], ops.signers[i] = ops.signers[i], ops.signers[j]
}

func sortOperationsWithSigners(ops []*Operation, signers [][]*crypto.PrivateKeySECP256K1R, codec codec.Manager) {
	sort.Sort(&innerSortOperationsWithSigners{ops: ops, signers: signers, codec: codec})
}
