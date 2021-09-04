// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"sort"

	"github.com/lasthyphen/paaro/ids"
	"github.com/lasthyphen/paaro/snow/choices"
)

// TestBlock is a useful test block
type TestBlock struct {
	choices.TestDecidable

	ParentV ids.ID
	HeightV uint64
	VerifyV error
	BytesV  []byte
}

// Parent implements the Block interface
func (b *TestBlock) Parent() ids.ID { return b.ParentV }

// Height returns the height of the block
func (b *TestBlock) Height() uint64 { return b.HeightV }

// Verify implements the Block interface
func (b *TestBlock) Verify() error { return b.VerifyV }

// Bytes implements the Block interface
func (b *TestBlock) Bytes() []byte { return b.BytesV }

type sortBlocks []*TestBlock

func (sb sortBlocks) Less(i, j int) bool { return sb[i].HeightV < sb[j].HeightV }
func (sb sortBlocks) Len() int           { return len(sb) }
func (sb sortBlocks) Swap(i, j int)      { sb[j], sb[i] = sb[i], sb[j] }

// SortTestBlocks sorts the array of blocks by height
func SortTestBlocks(blocks []*TestBlock) { sort.Sort(sortBlocks(blocks)) }
