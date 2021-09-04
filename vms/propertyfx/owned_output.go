package propertyfx

import (
	"github.com/lasthyphen/paaro/vms/secp256k1fx"
)

type OwnedOutput struct {
	secp256k1fx.OutputOwners `serialize:"true"`
}
