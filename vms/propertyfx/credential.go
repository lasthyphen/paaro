package propertyfx

import (
	"github.com/lasthyphen/paaro/vms/secp256k1fx"
)

type Credential struct {
	secp256k1fx.Credential `serialize:"true"`
}
