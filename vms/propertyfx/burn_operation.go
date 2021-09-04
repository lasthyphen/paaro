package propertyfx

import (
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/vms/components/verify"
	"github.com/djt-labs/paaro/vms/secp256k1fx"
)

type BurnOperation struct {
	secp256k1fx.Input `serialize:"true"`
}

func (op *BurnOperation) InitCtx(ctx *snow.Context) {}

func (op *BurnOperation) Outs() []verify.State { return nil }
