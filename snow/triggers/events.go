// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package triggers

import (
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
)

// Acceptor is implemented when a struct is monitoring if a message is accepted
type Acceptor interface {
	Accept(ctx *snow.Context, containerID ids.ID, container []byte) error
}

// Rejector is implemented when a struct is monitoring if a message is rejected
type Rejector interface {
	Reject(ctx *snow.Context, containerID ids.ID, container []byte) error
}

// Issuer is implemented when a struct is monitoring if a message is issued
type Issuer interface {
	Issue(ctx *snow.Context, containerID ids.ID, container []byte) error
}
