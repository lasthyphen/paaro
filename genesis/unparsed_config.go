// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/hex"
	"errors"

	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/utils/constants"
	"github.com/djt-labs/paaro/utils/formatting"
)

type UnparsedAllocation struct {
	ETHAddr        string         `json:"ethAddr"`
	DJTXAddr       string         `json:"djtxAddr"`
	InitialAmount  uint64         `json:"initialAmount"`
	UnlockSchedule []LockedAmount `json:"unlockSchedule"`
}

func (ua UnparsedAllocation) Parse() (Allocation, error) {
	a := Allocation{
		InitialAmount:  ua.InitialAmount,
		UnlockSchedule: ua.UnlockSchedule,
	}

	if len(ua.ETHAddr) < 2 {
		return a, errors.New("invalid eth address")
	}

	ethAddrBytes, err := hex.DecodeString(ua.ETHAddr[2:])
	if err != nil {
		return a, err
	}
	ethAddr, err := ids.ToShortID(ethAddrBytes)
	if err != nil {
		return a, err
	}
	a.ETHAddr = ethAddr

	_, _, djtxAddrBytes, err := formatting.ParseAddress(ua.DJTXAddr)
	if err != nil {
		return a, err
	}
	djtxAddr, err := ids.ToShortID(djtxAddrBytes)
	if err != nil {
		return a, err
	}
	a.DJTXAddr = djtxAddr

	return a, nil
}

type UnparsedStaker struct {
	NodeID        string `json:"nodeID"`
	RewardAddress string `json:"rewardAddress"`
	DelegationFee uint32 `json:"delegationFee"`
}

func (us UnparsedStaker) Parse() (Staker, error) {
	s := Staker{
		DelegationFee: us.DelegationFee,
	}

	nodeID, err := ids.ShortFromPrefixedString(us.NodeID, constants.NodeIDPrefix)
	if err != nil {
		return s, err
	}
	s.NodeID = nodeID

	_, _, djtxAddrBytes, err := formatting.ParseAddress(us.RewardAddress)
	if err != nil {
		return s, err
	}
	djtxAddr, err := ids.ToShortID(djtxAddrBytes)
	if err != nil {
		return s, err
	}
	s.RewardAddress = djtxAddr
	return s, nil
}

// UnparsedConfig contains the genesis addresses used to construct a genesis
type UnparsedConfig struct {
	NetworkID uint32 `json:"networkID"`

	Allocations []UnparsedAllocation `json:"allocations"`

	StartTime                  uint64           `json:"startTime"`
	InitialStakeDuration       uint64           `json:"initialStakeDuration"`
	InitialStakeDurationOffset uint64           `json:"initialStakeDurationOffset"`
	InitialStakedFunds         []string         `json:"initialStakedFunds"`
	InitialStakers             []UnparsedStaker `json:"initialStakers"`

	CChainGenesis string `json:"cChainGenesis"`

	Message string `json:"message"`
}

func (uc UnparsedConfig) Parse() (Config, error) {
	c := Config{
		NetworkID:                  uc.NetworkID,
		Allocations:                make([]Allocation, len(uc.Allocations)),
		StartTime:                  uc.StartTime,
		InitialStakeDuration:       uc.InitialStakeDuration,
		InitialStakeDurationOffset: uc.InitialStakeDurationOffset,
		InitialStakedFunds:         make([]ids.ShortID, len(uc.InitialStakedFunds)),
		InitialStakers:             make([]Staker, len(uc.InitialStakers)),
		CChainGenesis:              uc.CChainGenesis,
		Message:                    uc.Message,
	}
	for i, ua := range uc.Allocations {
		a, err := ua.Parse()
		if err != nil {
			return c, err
		}
		c.Allocations[i] = a
	}
	for i, isa := range uc.InitialStakedFunds {
		_, _, djtxAddrBytes, err := formatting.ParseAddress(isa)
		if err != nil {
			return c, err
		}
		djtxAddr, err := ids.ToShortID(djtxAddrBytes)
		if err != nil {
			return c, err
		}
		c.InitialStakedFunds[i] = djtxAddr
	}
	for i, uis := range uc.InitialStakers {
		is, err := uis.Parse()
		if err != nil {
			return c, err
		}
		c.InitialStakers[i] = is
	}
	return c, nil
}
