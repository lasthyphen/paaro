// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/djt-labs/paaro/codec"
	"github.com/djt-labs/paaro/codec/linearcodec"
	"github.com/djt-labs/paaro/codec/reflectcodec"
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/utils/constants"
	"github.com/djt-labs/paaro/utils/formatting"
	"github.com/djt-labs/paaro/utils/json"
	"github.com/djt-labs/paaro/utils/wrappers"
	"github.com/djt-labs/paaro/vms/dvm"
	"github.com/djt-labs/paaro/vms/evm"
	"github.com/djt-labs/paaro/vms/nftfx"
	"github.com/djt-labs/paaro/vms/platformvm"
	"github.com/djt-labs/paaro/vms/propertyfx"
	"github.com/djt-labs/paaro/vms/secp256k1fx"
)

const (
	defaultEncoding    = formatting.Hex
	codecVersion       = 0
	configChainIDAlias = "X"
)

// validateInitialStakedFunds ensures all staked
// funds have allocations and that all staked
// funds are unique.
//
// This function assumes that NetworkID in *Config has already
// been checked for correctness.
func validateInitialStakedFunds(config *Config) error {
	if len(config.InitialStakedFunds) == 0 {
		return errors.New("initial staked funds cannot be empty")
	}

	allocationSet := ids.ShortSet{}
	initialStakedFundsSet := ids.ShortSet{}
	for _, allocation := range config.Allocations {
		// It is ok to have duplicates as different
		// ethAddrs could claim to the same djtxAddr.
		allocationSet.Add(allocation.DJTXAddr)
	}

	for _, staker := range config.InitialStakedFunds {
		if initialStakedFundsSet.Contains(staker) {
			djtxAddr, err := formatting.FormatAddress(
				configChainIDAlias,
				constants.GetHRP(config.NetworkID),
				staker.Bytes(),
			)
			if err != nil {
				return fmt.Errorf(
					"unable to format address from %s",
					staker.String(),
				)
			}

			return fmt.Errorf(
				"address %s is duplicated in initial staked funds",
				djtxAddr,
			)
		}
		initialStakedFundsSet.Add(staker)

		if !allocationSet.Contains(staker) {
			djtxAddr, err := formatting.FormatAddress(
				configChainIDAlias,
				constants.GetHRP(config.NetworkID),
				staker.Bytes(),
			)
			if err != nil {
				return fmt.Errorf(
					"unable to format address from %s",
					staker.String(),
				)
			}

			return fmt.Errorf(
				"address %s does not have an allocation to stake",
				djtxAddr,
			)
		}
	}

	return nil
}

// validateConfig returns an error if the provided
// *Config is not considered valid.
func validateConfig(networkID uint32, config *Config) error {
	if networkID != config.NetworkID {
		return fmt.Errorf(
			"networkID %d specified but genesis config contains networkID %d",
			networkID,
			config.NetworkID,
		)
	}

	initialSupply, err := config.InitialSupply()
	switch {
	case err != nil:
		return fmt.Errorf("unable to calculate initial supply: %w", err)
	case initialSupply == 0:
		return errors.New("initial supply must be > 0")
	}

	startTime := time.Unix(int64(config.StartTime), 0)
	if time.Since(startTime) < 0 {
		return fmt.Errorf(
			"start time cannot be in the future: %s",
			startTime,
		)
	}

	// We don't impose any restrictions on the minimum
	// stake duration to enable complex testing configurations
	// but recommend setting a minimum duration of at least
	// 15 minutes.
	if config.InitialStakeDuration == 0 {
		return errors.New("initial stake duration must be > 0")
	}

	if len(config.InitialStakers) == 0 {
		return errors.New("initial stakers must be > 0")
	}

	offsetTimeRequired := config.InitialStakeDurationOffset * uint64(len(config.InitialStakers)-1)
	if offsetTimeRequired > config.InitialStakeDuration {
		return fmt.Errorf(
			"initial stake duration is %d but need at least %d with offset of %d",
			config.InitialStakeDuration,
			offsetTimeRequired,
			config.InitialStakeDurationOffset,
		)
	}

	if err := validateInitialStakedFunds(config); err != nil {
		return fmt.Errorf("initial staked funds validation failed: %w", err)
	}

	if len(config.CChainGenesis) == 0 {
		return errors.New("C-Chain genesis cannot be empty")
	}

	return nil
}

// Genesis returns the genesis data of the Platform Chain.
//
// Since an Dijets network has exactly one Platform Chain, and the Platform
// Chain defines the genesis state of the network (who is staking, which chains
// exist, etc.), defining the genesis state of the Platform Chain is the same as
// defining the genesis state of the network.
//
// Genesis accepts:
// 1) The ID of the new network. [networkID]
// 2) The location of a custom genesis config to load. [filepath]
//
// If [filepath] is empty or the given network ID is Mainnet, Testnet, or Local, loads the
// network genesis state from predefined configs. If [filepath] is non-empty and networkID
// isn't Mainnet, Testnet, or Local, loads the network genesis data from the config at [filepath].
//
// Genesis returns:
// 1) The byte representation of the genesis state of the platform chain
//    (ie the genesis state of the network)
// 2) The asset ID of DJTX
func Genesis(networkID uint32, filepath string) ([]byte, ids.ID, error) {
	config := GetConfig(networkID)
	if len(filepath) > 0 {
		switch networkID {
		case constants.MainnetID, constants.TestnetID, constants.LocalID:
			return nil, ids.ID{}, fmt.Errorf(
				"cannot override genesis config for standard network %s (%d)",
				constants.NetworkName(networkID),
				networkID,
			)
		}

		customConfig, err := GetConfigFile(filepath)
		if err != nil {
			return nil, ids.ID{}, fmt.Errorf("unable to load provided genesis config at %s: %w", filepath, err)
		}

		config = customConfig
	}

	if err := validateConfig(networkID, config); err != nil {
		return nil, ids.ID{}, fmt.Errorf("genesis config validation failed: %w", err)
	}

	return FromConfig(config)
}

// FromConfig returns:
// 1) The byte representation of the genesis state of the platform chain
//    (ie the genesis state of the network)
// 2) The asset ID of DJTX
func FromConfig(config *Config) ([]byte, ids.ID, error) {
	hrp := constants.GetHRP(config.NetworkID)

	amount := uint64(0)

	// Specify the genesis state of the DVM
	dvmArgs := dvm.BuildGenesisArgs{
		NetworkID: json.Uint32(config.NetworkID),
		Encoding:  defaultEncoding,
	}
	{
		djtx := dvm.AssetDefinition{
			Name:         "Dijets",
			Symbol:       "DJTX",
			Denomination: 9,
			InitialState: map[string][]interface{}{},
		}
		memoBytes := []byte{}
		xAllocations := []Allocation(nil)
		for _, allocation := range config.Allocations {
			if allocation.InitialAmount > 0 {
				xAllocations = append(xAllocations, allocation)
			}
		}
		sortXAllocation(xAllocations)

		for _, allocation := range xAllocations {
			addr, err := formatting.FormatBech32(hrp, allocation.DJTXAddr.Bytes())
			if err != nil {
				return nil, ids.ID{}, err
			}

			djtx.InitialState["fixedCap"] = append(djtx.InitialState["fixedCap"], dvm.Holder{
				Amount:  json.Uint64(allocation.InitialAmount),
				Address: addr,
			})
			memoBytes = append(memoBytes, allocation.ETHAddr.Bytes()...)
			amount += allocation.InitialAmount
		}

		var err error
		djtx.Memo, err = formatting.EncodeWithChecksum(defaultEncoding, memoBytes)
		if err != nil {
			return nil, ids.Empty, fmt.Errorf("couldn't parse memo bytes to string: %w", err)
		}
		dvmArgs.GenesisData = map[string]dvm.AssetDefinition{
			"DJTX": djtx, // The DVM starts out with one asset: DJTX
		}
	}
	dvmReply := dvm.BuildGenesisReply{}

	dvmSS := dvm.CreateStaticService()
	err := dvmSS.BuildGenesis(nil, &dvmArgs, &dvmReply)
	if err != nil {
		return nil, ids.ID{}, err
	}

	bytes, err := formatting.Decode(defaultEncoding, dvmReply.Bytes)
	if err != nil {
		return nil, ids.ID{}, fmt.Errorf("couldn't parse dvm genesis reply: %w", err)
	}
	djtxAssetID, err := DJTXAssetID(bytes)
	if err != nil {
		return nil, ids.ID{}, fmt.Errorf("couldn't generate DJTX asset ID: %w", err)
	}

	genesisTime := time.Unix(int64(config.StartTime), 0)
	initialSupply, err := config.InitialSupply()
	if err != nil {
		return nil, ids.ID{}, fmt.Errorf("couldn't calculate the initial supply: %w", err)
	}

	initiallyStaked := ids.ShortSet{}
	initiallyStaked.Add(config.InitialStakedFunds...)
	skippedAllocations := []Allocation(nil)

	// Specify the initial state of the Platform Chain
	platformvmArgs := platformvm.BuildGenesisArgs{
		DjtxAssetID:   djtxAssetID,
		NetworkID:     json.Uint32(config.NetworkID),
		Time:          json.Uint64(config.StartTime),
		InitialSupply: json.Uint64(initialSupply),
		Message:       config.Message,
		Encoding:      defaultEncoding,
	}
	for _, allocation := range config.Allocations {
		if initiallyStaked.Contains(allocation.DJTXAddr) {
			skippedAllocations = append(skippedAllocations, allocation)
			continue
		}
		addr, err := formatting.FormatBech32(hrp, allocation.DJTXAddr.Bytes())
		if err != nil {
			return nil, ids.ID{}, err
		}
		for _, unlock := range allocation.UnlockSchedule {
			if unlock.Amount > 0 {
				msgStr, err := formatting.EncodeWithChecksum(defaultEncoding, allocation.ETHAddr.Bytes())
				if err != nil {
					return nil, ids.Empty, fmt.Errorf("couldn't encode message: %w", err)
				}
				platformvmArgs.UTXOs = append(platformvmArgs.UTXOs,
					platformvm.APIUTXO{
						Locktime: json.Uint64(unlock.Locktime),
						Amount:   json.Uint64(unlock.Amount),
						Address:  addr,
						Message:  msgStr,
					},
				)
				amount += unlock.Amount
			}
		}
	}

	allNodeAllocations := splitAllocations(skippedAllocations, len(config.InitialStakers))
	endStakingTime := genesisTime.Add(time.Duration(config.InitialStakeDuration) * time.Second)
	stakingOffset := time.Duration(0)
	for i, staker := range config.InitialStakers {
		nodeAllocations := allNodeAllocations[i]
		endStakingTime := endStakingTime.Add(-stakingOffset)
		stakingOffset += time.Duration(config.InitialStakeDurationOffset) * time.Second

		destAddrStr, err := formatting.FormatBech32(hrp, staker.RewardAddress.Bytes())
		if err != nil {
			return nil, ids.ID{}, err
		}

		utxos := []platformvm.APIUTXO(nil)
		for _, allocation := range nodeAllocations {
			addr, err := formatting.FormatBech32(hrp, allocation.DJTXAddr.Bytes())
			if err != nil {
				return nil, ids.ID{}, err
			}
			for _, unlock := range allocation.UnlockSchedule {
				msgStr, err := formatting.EncodeWithChecksum(defaultEncoding, allocation.ETHAddr.Bytes())
				if err != nil {
					return nil, ids.Empty, fmt.Errorf("couldn't encode message: %w", err)
				}
				utxos = append(utxos, platformvm.APIUTXO{
					Locktime: json.Uint64(unlock.Locktime),
					Amount:   json.Uint64(unlock.Amount),
					Address:  addr,
					Message:  msgStr,
				})
				amount += unlock.Amount
			}
		}

		delegationFee := json.Uint32(staker.DelegationFee)

		platformvmArgs.Validators = append(platformvmArgs.Validators,
			platformvm.APIPrimaryValidator{
				APIStaker: platformvm.APIStaker{
					StartTime: json.Uint64(genesisTime.Unix()),
					EndTime:   json.Uint64(endStakingTime.Unix()),
					NodeID:    staker.NodeID.PrefixedString(constants.NodeIDPrefix),
				},
				RewardOwner: &platformvm.APIOwner{
					Threshold: 1,
					Addresses: []string{destAddrStr},
				},
				Staked:             utxos,
				ExactDelegationFee: &delegationFee,
			},
		)
	}

	// Specify the chains that exist upon this network's creation
	genesisStr, err := formatting.EncodeWithChecksum(defaultEncoding, []byte(config.CChainGenesis))
	if err != nil {
		return nil, ids.Empty, fmt.Errorf("couldn't encode message: %w", err)
	}
	platformvmArgs.Chains = []platformvm.APIChain{
		{
			GenesisData: dvmReply.Bytes,
			SubnetID:    constants.PrimaryNetworkID,
			VMID:        dvm.ID,
			FxIDs: []ids.ID{
				secp256k1fx.ID,
				nftfx.ID,
				propertyfx.ID,
			},
			Name: "X-Chain",
		},
		{
			GenesisData: genesisStr,
			SubnetID:    constants.PrimaryNetworkID,
			VMID:        evm.ID,
			Name:        "C-Chain",
		},
	}

	platformvmReply := platformvm.BuildGenesisReply{}
	platformvmSS := platformvm.StaticService{}
	if err := platformvmSS.BuildGenesis(nil, &platformvmArgs, &platformvmReply); err != nil {
		return nil, ids.ID{}, fmt.Errorf("problem while building platform chain's genesis state: %w", err)
	}

	genesisBytes, err := formatting.Decode(platformvmReply.Encoding, platformvmReply.Bytes)
	if err != nil {
		return nil, ids.ID{}, fmt.Errorf("problem parsing platformvm genesis bytes: %w", err)
	}

	return genesisBytes, djtxAssetID, nil
}

func splitAllocations(allocations []Allocation, numSplits int) [][]Allocation {
	totalAmount := uint64(0)
	for _, allocation := range allocations {
		for _, unlock := range allocation.UnlockSchedule {
			totalAmount += unlock.Amount
		}
	}

	nodeWeight := totalAmount / uint64(numSplits)
	allNodeAllocations := make([][]Allocation, 0, numSplits)

	currentNodeAllocation := []Allocation(nil)
	currentNodeAmount := uint64(0)
	for _, allocation := range allocations {
		currentAllocation := allocation
		// Already added to the X-chain
		currentAllocation.InitialAmount = 0
		// Going to be added until the correct amount is reached
		currentAllocation.UnlockSchedule = nil

		for _, unlock := range allocation.UnlockSchedule {
			unlock := unlock
			for currentNodeAmount+unlock.Amount > nodeWeight && len(allNodeAllocations) < numSplits-1 {
				amountToAdd := nodeWeight - currentNodeAmount
				currentAllocation.UnlockSchedule = append(currentAllocation.UnlockSchedule, LockedAmount{
					Amount:   amountToAdd,
					Locktime: unlock.Locktime,
				})
				unlock.Amount -= amountToAdd

				currentNodeAllocation = append(currentNodeAllocation, currentAllocation)

				allNodeAllocations = append(allNodeAllocations, currentNodeAllocation)

				currentNodeAllocation = nil
				currentNodeAmount = 0

				currentAllocation = allocation
				// Already added to the X-chain
				currentAllocation.InitialAmount = 0
				// Going to be added until the correct amount is reached
				currentAllocation.UnlockSchedule = nil
			}

			if unlock.Amount == 0 {
				continue
			}

			currentAllocation.UnlockSchedule = append(currentAllocation.UnlockSchedule, LockedAmount{
				Amount:   unlock.Amount,
				Locktime: unlock.Locktime,
			})
			currentNodeAmount += unlock.Amount
		}

		if len(currentAllocation.UnlockSchedule) > 0 {
			currentNodeAllocation = append(currentNodeAllocation, currentAllocation)
		}
	}

	return append(allNodeAllocations, currentNodeAllocation)
}

func VMGenesis(genesisBytes []byte, vmID ids.ID) (*platformvm.Tx, error) {
	genesis := platformvm.Genesis{}
	if _, err := platformvm.GenesisCodec.Unmarshal(genesisBytes, &genesis); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal genesis bytes due to: %w", err)
	}
	if err := genesis.Initialize(); err != nil {
		return nil, err
	}
	for _, chain := range genesis.Chains {
		uChain := chain.UnsignedTx.(*platformvm.UnsignedCreateChainTx)
		if uChain.VMID == vmID {
			return chain, nil
		}
	}
	return nil, fmt.Errorf("couldn't find blockchain with VM ID %s", vmID)
}

func DJTXAssetID(dvmGenesisBytes []byte) (ids.ID, error) {
	c := linearcodec.New(reflectcodec.DefaultTagName, 1<<20)
	m := codec.NewManager(math.MaxUint32)
	errs := wrappers.Errs{}
	errs.Add(
		c.RegisterType(&dvm.BaseTx{}),
		c.RegisterType(&dvm.CreateAssetTx{}),
		c.RegisterType(&dvm.OperationTx{}),
		c.RegisterType(&dvm.ImportTx{}),
		c.RegisterType(&dvm.ExportTx{}),
		c.RegisterType(&secp256k1fx.TransferInput{}),
		c.RegisterType(&secp256k1fx.MintOutput{}),
		c.RegisterType(&secp256k1fx.TransferOutput{}),
		c.RegisterType(&secp256k1fx.MintOperation{}),
		c.RegisterType(&secp256k1fx.Credential{}),
		m.RegisterCodec(codecVersion, c),
	)
	if errs.Errored() {
		return ids.ID{}, errs.Err
	}

	genesis := dvm.Genesis{}
	if _, err := m.Unmarshal(dvmGenesisBytes, &genesis); err != nil {
		return ids.ID{}, err
	}

	if len(genesis.Txs) == 0 {
		return ids.ID{}, errors.New("genesis creates no transactions")
	}
	genesisTx := genesis.Txs[0]

	tx := dvm.Tx{UnsignedTx: &genesisTx.CreateAssetTx}
	unsignedBytes, err := m.Marshal(codecVersion, tx.UnsignedTx)
	if err != nil {
		return ids.ID{}, err
	}
	signedBytes, err := m.Marshal(codecVersion, &tx)
	if err != nil {
		return ids.ID{}, err
	}
	tx.Initialize(unsignedBytes, signedBytes)

	return tx.ID(), nil
}

type innerSortXAllocation []Allocation

func (xa innerSortXAllocation) Less(i, j int) bool {
	return xa[i].InitialAmount < xa[j].InitialAmount ||
		(xa[i].InitialAmount == xa[j].InitialAmount &&
			bytes.Compare(xa[i].DJTXAddr.Bytes(), xa[j].DJTXAddr.Bytes()) == -1)
}

func (xa innerSortXAllocation) Len() int      { return len(xa) }
func (xa innerSortXAllocation) Swap(i, j int) { xa[j], xa[i] = xa[i], xa[j] }

func sortXAllocation(a []Allocation) { sort.Sort(innerSortXAllocation(a)) }
