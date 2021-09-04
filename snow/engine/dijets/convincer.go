// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dijets

import (
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow/consensus/dijets"
	"github.com/djt-labs/paaro/snow/engine/common"
	"github.com/djt-labs/paaro/utils/wrappers"
)

// convincer sends chits to [vdr] after its dependencies are met.
type convincer struct {
	consensus dijets.Consensus
	sender    common.Sender
	vdr       ids.ShortID
	requestID uint32
	sent      bool
	abandoned bool
	deps      ids.Set
	errs      *wrappers.Errs
}

func (c *convincer) Dependencies() ids.Set { return c.deps }

// Mark that a dependency has been met.
func (c *convincer) Fulfill(id ids.ID) {
	c.deps.Remove(id)
	c.Update()
}

// Abandon this attempt to send chits.
func (c *convincer) Abandon(ids.ID) {
	c.abandoned = true
	c.Update()
}

func (c *convincer) Update() {
	if c.sent || c.errs.Errored() || (!c.abandoned && c.deps.Len() != 0) {
		return
	}
	c.sent = true

	c.sender.Chits(c.vdr, c.requestID, c.consensus.Preferences().List())
}
