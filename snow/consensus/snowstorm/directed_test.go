// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowstorm

import (
	"testing"
)

func TestDirectedConsensus(t *testing.T) { ConsensusTest(t, DirectedFactory{}, "DG") }
