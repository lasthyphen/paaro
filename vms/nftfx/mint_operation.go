package nftfx

import (
	"errors"

	"github.com/lasthyphen/paaro/vms/types"

	"github.com/lasthyphen/paaro/snow"
	"github.com/lasthyphen/paaro/vms/components/verify"
	"github.com/lasthyphen/paaro/vms/secp256k1fx"
)

var errNilMintOperation = errors.New("nil mint operation")

type MintOperation struct {
	MintInput secp256k1fx.Input           `serialize:"true" json:"mintInput"`
	GroupID   uint32                      `serialize:"true" json:"groupID"`
	Payload   types.JSONByteSlice         `serialize:"true" json:"payload"`
	Outputs   []*secp256k1fx.OutputOwners `serialize:"true" json:"outputs"`
}

func (op *MintOperation) InitCtx(ctx *snow.Context) {
	for _, out := range op.Outputs {
		out.InitCtx(ctx)
	}
}

func (op *MintOperation) Cost() (uint64, error) {
	return op.MintInput.Cost()
}

// Outs Returns []TransferOutput as []verify.State
func (op *MintOperation) Outs() []verify.State {
	outs := []verify.State{}
	for _, out := range op.Outputs {
		outs = append(outs, &TransferOutput{
			GroupID:      op.GroupID,
			Payload:      op.Payload,
			OutputOwners: *out,
		})
	}
	return outs
}

func (op *MintOperation) Verify() error {
	switch {
	case op == nil:
		return errNilMintOperation
	case len(op.Payload) > MaxPayloadSize:
		return errPayloadTooLarge
	}

	for _, out := range op.Outputs {
		if err := out.Verify(); err != nil {
			return err
		}
	}
	return op.MintInput.Verify()
}
