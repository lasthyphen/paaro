// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/djt-labs/paaro/database/memdb"
	"github.com/djt-labs/paaro/snow/consensus/snowball"
	"github.com/djt-labs/paaro/snow/consensus/snowman"
	"github.com/djt-labs/paaro/snow/engine/common"
	"github.com/djt-labs/paaro/snow/engine/common/queue"
	"github.com/djt-labs/paaro/snow/engine/snowman/block"
	"github.com/djt-labs/paaro/snow/engine/snowman/bootstrap"
)

func DefaultConfig() Config {
	blocked, _ := queue.NewWithMissing(memdb.New(), "", prometheus.NewRegistry())
	return Config{
		Config: bootstrap.Config{
			Config:  common.DefaultConfigTest(),
			Blocked: blocked,
			VM:      &block.TestVM{},
		},
		Params: snowball.Parameters{
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
		Consensus: &snowman.Topological{},
	}
}
