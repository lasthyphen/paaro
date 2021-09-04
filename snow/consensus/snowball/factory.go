// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowball

// Factory returns new instances of Consensus
type Factory interface {
	New() Consensus
}
