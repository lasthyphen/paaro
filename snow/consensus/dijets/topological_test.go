// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dijets

import (
	"testing"
)

func TestTopological(t *testing.T) { ConsensusTest(t, TopologicalFactory{}) }
