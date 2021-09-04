// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow/consensus/snowman"
)

// issuer issues [blk] into to consensus after its dependencies are met.
type issuer struct {
	t         *Transitive
	blk       snowman.Block
	abandoned bool
	deps      ids.Set
}

func (i *issuer) Dependencies() ids.Set { return i.deps }

// Mark that a dependency has been met
func (i *issuer) Fulfill(id ids.ID) {
	i.deps.Remove(id)
	i.Update()
}

// Abandon the attempt to issue [i.block]
func (i *issuer) Abandon(ids.ID) {
	if !i.abandoned {
		blkID := i.blk.ID()
		i.t.pending.Remove(blkID)
		i.t.blocked.Abandon(blkID)

		// Tracks performance statistics
		i.t.metrics.numRequests.Set(float64(i.t.blkReqs.Len()))
		i.t.metrics.numBlocked.Set(float64(i.t.pending.Len()))
		i.t.metrics.numBlockers.Set(float64(i.t.blocked.Len()))
	}
	i.abandoned = true
}

func (i *issuer) Update() {
	if i.abandoned || i.deps.Len() != 0 || i.t.errs.Errored() {
		return
	}
	// Issue the block into consensus
	i.t.errs.Add(i.t.deliver(i.blk))
}
