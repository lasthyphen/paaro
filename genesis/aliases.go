// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/utils/constants"
	"github.com/djt-labs/paaro/vms/dvm"
	"github.com/djt-labs/paaro/vms/evm"
	"github.com/djt-labs/paaro/vms/nftfx"
	"github.com/djt-labs/paaro/vms/platformvm"
	"github.com/djt-labs/paaro/vms/propertyfx"
	"github.com/djt-labs/paaro/vms/secp256k1fx"
)

// Aliases returns the default aliases based on the network ID
func Aliases(genesisBytes []byte) (map[string][]string, map[ids.ID][]string, error) {
	apiAliases := getAPIAliases()
	chainAliases := map[ids.ID][]string{
		constants.PlatformChainID: {"P", "platform"},
	}
	genesis := &platformvm.Genesis{} // TODO let's not re-create genesis to do aliasing
	if _, err := platformvm.GenesisCodec.Unmarshal(genesisBytes, genesis); err != nil {
		return nil, nil, err
	}
	if err := genesis.Initialize(); err != nil {
		return nil, nil, err
	}

	for _, chain := range genesis.Chains {
		uChain := chain.UnsignedTx.(*platformvm.UnsignedCreateChainTx)
		switch uChain.VMID {
		case dvm.ID:
			apiAliases[constants.ChainAliasPrefix+chain.ID().String()] = []string{"X", "dvm", constants.ChainAliasPrefix + "X", constants.ChainAliasPrefix + "/dvm"}
			chainAliases[chain.ID()] = GetXChainAliases()
		case evm.ID:
			apiAliases[constants.ChainAliasPrefix+chain.ID().String()] = []string{"C", "evm", constants.ChainAliasPrefix + "C", constants.ChainAliasPrefix + "evm"}
			chainAliases[chain.ID()] = GetCChainAliases()
		}
	}
	return apiAliases, chainAliases, nil
}

func GetCChainAliases() []string {
	return []string{"C", "evm"}
}

func GetXChainAliases() []string {
	return []string{"X", "dvm"}
}

func getAPIAliases() map[string][]string {
	return map[string][]string{
		constants.VMAliasPrefix + platformvm.ID.String():                {constants.VMAliasPrefix + "platform"},
		constants.VMAliasPrefix + dvm.ID.String():                       {constants.VMAliasPrefix + "dvm"},
		constants.VMAliasPrefix + evm.ID.String():                       {constants.VMAliasPrefix + "evm"},
		constants.ChainAliasPrefix + constants.PlatformChainID.String(): {"P", "platform", constants.ChainAliasPrefix + "P", constants.ChainAliasPrefix + "platform"},
	}
}

func GetVMAliases() map[ids.ID][]string {
	return map[ids.ID][]string{
		platformvm.ID:  {"platform"},
		dvm.ID:         {"dvm"},
		evm.ID:         {"evm"},
		secp256k1fx.ID: {"secp256k1fx"},
		nftfx.ID:       {"nftfx"},
		propertyfx.ID:  {"propertyfx"},
	}
}
