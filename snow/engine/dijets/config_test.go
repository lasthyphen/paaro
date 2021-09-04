// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dijets

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/djt-labs/paaro/database/memdb"
	"github.com/djt-labs/paaro/snow/consensus/dijets"
	"github.com/djt-labs/paaro/snow/consensus/snowball"
	"github.com/djt-labs/paaro/snow/engine/dijets/bootstrap"
	"github.com/djt-labs/paaro/snow/engine/dijets/vertex"
	"github.com/djt-labs/paaro/snow/engine/common"
	"github.com/djt-labs/paaro/snow/engine/common/queue"
)

func DefaultConfig() Config {
	vtxBlocked, _ := queue.NewWithMissing(memdb.New(), "", prometheus.NewRegistry())
	txBlocked, _ := queue.New(memdb.New(), "", prometheus.NewRegistry())
	return Config{
		Config: bootstrap.Config{
			Config:     common.DefaultConfigTest(),
			VtxBlocked: vtxBlocked,
			TxBlocked:  txBlocked,
			Manager:    &vertex.TestManager{},
			VM:         &vertex.TestVM{},
		},
		Params: dijets.Parameters{
			Parameters: snowball.Parameters{
				Metrics:               prometheus.NewRegistry(),
				K:                     1,
				Alpha:                 1,
				BetaVirtuous:          1,
				BetaRogue:             2,
				ConcurrentRepolls:     1,
				OptimalProcessing:     100,
				MaxOutstandingItems:   1,
				MaxItemProcessingTime: 1,
			},
			Parents:   2,
			BatchSize: 1,
		},
		Consensus: &dijets.Topological{},
	}
}
