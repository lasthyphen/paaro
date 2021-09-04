// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package crypto

import (
	"errors"
)

var (
	errInvalidSigLen = errors.New("invalid signature length")
	errMutatedSig    = errors.New("signature was mutated from its original format")
)
