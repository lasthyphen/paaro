// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package common

import (
	"github.com/lasthyphen/paaro/ids"
)

// Bootstrapable defines the functionality required to support bootstrapping
type Bootstrapable interface {
	// Returns the set of containerIDs that are accepted, but have no accepted
	// children.
	CurrentAcceptedFrontier() ([]ids.ID, error)

	// Returns the subset of containerIDs that are accepted by this chain.
	FilterAccepted(containerIDs []ids.ID) (acceptedContainerIDs []ids.ID)

	// Force the provided containers to be accepted. Only returns fatal errors
	// if they occur.
	ForceAccepted(acceptedContainerIDs []ids.ID) error
}
