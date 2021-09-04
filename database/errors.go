// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package database

import "errors"

// common errors
var (
	ErrClosed          = errors.New("closed")
	ErrNotFound        = errors.New("not found")
	ErrAvoidCorruption = errors.New("closed to avoid possible corruption")
)
