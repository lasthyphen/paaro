package dvm

import (
	"bytes"
	"testing"

	"github.com/djt-labs/paaro/pubsub"
	"github.com/djt-labs/paaro/vms/components/djtx"
	"github.com/djt-labs/paaro/vms/secp256k1fx"
	"github.com/stretchr/testify/assert"

	"github.com/djt-labs/paaro/ids"
)

type mockFilter struct {
	addr []byte
}

func (f *mockFilter) Check(addr []byte) bool {
	return bytes.Equal(addr, f.addr)
}

func TestFilter(t *testing.T) {
	assert := assert.New(t)

	addrID := ids.ShortID{1}
	tx := Tx{UnsignedTx: &BaseTx{BaseTx: djtx.BaseTx{
		Outs: []*djtx.TransferableOutput{
			{
				Out: &secp256k1fx.TransferOutput{
					OutputOwners: secp256k1fx.OutputOwners{
						Addrs: []ids.ShortID{addrID},
					},
				},
			},
		},
	}}}
	addrBytes := addrID[:]

	fp := pubsub.NewFilterParam()
	err := fp.Add(addrBytes)
	assert.NoError(err)

	parser := NewPubSubFilterer(&tx)
	fr, _ := parser.Filter([]pubsub.Filter{&mockFilter{addr: addrBytes}})
	assert.Equal([]bool{true}, fr)
}
