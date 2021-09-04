// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package formatting

type CustomStringer struct{ Stringer func() string }

func (cs CustomStringer) String() string { return cs.Stringer() }
