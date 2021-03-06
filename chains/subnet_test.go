// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chains

import (
	"testing"

	"github.com/lasthyphen/paaro/ids"
	"github.com/stretchr/testify/assert"
)

func TestSubnet(t *testing.T) {
	assert := assert.New(t)

	chainID0 := ids.GenerateTestID()
	chainID1 := ids.GenerateTestID()
	chainID2 := ids.GenerateTestID()

	s := newSubnet(nil, chainID0)
	assert.False(s.IsBootstrapped(), "A subnet with one chain in bootstrapping shouldn't be considered bootstrapped")

	s.Bootstrapped(chainID0)
	assert.True(s.IsBootstrapped(), "A subnet with only bootstrapped chains should be considered bootstrapped")

	s.addChain(chainID1)
	assert.False(s.IsBootstrapped(), "A subnet with one chain in bootstrapping shouldn't be considered bootstrapped")

	s.addChain(chainID2)
	assert.False(s.IsBootstrapped(), "A subnet with one chain in bootstrapping shouldn't be considered bootstrapped")

	s.Bootstrapped(chainID1)
	assert.False(s.IsBootstrapped(), "A subnet with one chain in bootstrapping shouldn't be considered bootstrapped")

	s.Bootstrapped(chainID2)
	assert.True(s.IsBootstrapped(), "A subnet with only bootstrapped chains should be considered bootstrapped")
}
