package nftfx

import (
	"github.com/djt-labs/paaro/vms/secp256k1fx"
)

type Credential struct {
	secp256k1fx.Credential `serialize:"true"`
}
