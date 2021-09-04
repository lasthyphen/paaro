// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"math"

	"github.com/lasthyphen/paaro/codec"
	"github.com/lasthyphen/paaro/codec/linearcodec"
	"github.com/lasthyphen/paaro/codec/reflectcodec"
	"github.com/lasthyphen/paaro/utils/wrappers"
	"github.com/lasthyphen/paaro/vms/secp256k1fx"
)

const (
	codecVersion = 0
)

// Codecs do serialization and deserialization
var (
	Codec        codec.Manager
	GenesisCodec codec.Manager
)

func init() {
	c := linearcodec.NewDefault()
	Codec = codec.NewDefaultManager()
	gc := linearcodec.New(reflectcodec.DefaultTagName, math.MaxUint32)
	GenesisCodec = codec.NewManager(math.MaxUint32)

	errs := wrappers.Errs{}
	for _, c := range []codec.Registry{c, gc} {
		errs.Add(
			c.RegisterType(&ProposalBlock{}),
			c.RegisterType(&AbortBlock{}),
			c.RegisterType(&CommitBlock{}),
			c.RegisterType(&StandardBlock{}),
			c.RegisterType(&AtomicBlock{}),

			// The Fx is registered here because this is the same place it is
			// registered in the DVM. This ensures that the typeIDs match up for
			// utxos in shared memory.
			c.RegisterType(&secp256k1fx.TransferInput{}),
			c.RegisterType(&secp256k1fx.MintOutput{}),
			c.RegisterType(&secp256k1fx.TransferOutput{}),
			c.RegisterType(&secp256k1fx.MintOperation{}),
			c.RegisterType(&secp256k1fx.Credential{}),
			c.RegisterType(&secp256k1fx.Input{}),
			c.RegisterType(&secp256k1fx.OutputOwners{}),

			c.RegisterType(&UnsignedAddValidatorTx{}),
			c.RegisterType(&UnsignedAddSubnetValidatorTx{}),
			c.RegisterType(&UnsignedAddDelegatorTx{}),

			c.RegisterType(&UnsignedCreateChainTx{}),
			c.RegisterType(&UnsignedCreateSubnetTx{}),

			c.RegisterType(&UnsignedImportTx{}),
			c.RegisterType(&UnsignedExportTx{}),

			c.RegisterType(&UnsignedAdvanceTimeTx{}),
			c.RegisterType(&UnsignedRewardValidatorTx{}),

			c.RegisterType(&StakeableLockIn{}),
			c.RegisterType(&StakeableLockOut{}),
		)
	}
	errs.Add(
		Codec.RegisterCodec(codecVersion, c),
		GenesisCodec.RegisterCodec(codecVersion, gc),
	)
	if errs.Errored() {
		panic(errs.Err)
	}
}
