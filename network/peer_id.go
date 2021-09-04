// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"time"

	"github.com/lasthyphen/paaro/ids"
)

type PeerID struct {
	IP           string    `json:"ip"`
	PublicIP     string    `json:"publicIP"`
	ID           string    `json:"nodeID"`
	Version      string    `json:"version"`
	LastSent     time.Time `json:"lastSent"`
	LastReceived time.Time `json:"lastReceived"`
	Benched      []ids.ID  `json:"benched"`
}
