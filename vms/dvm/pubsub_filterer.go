package dvm

import (
	"github.com/djt-labs/paaro/api"
	"github.com/djt-labs/paaro/pubsub"
	"github.com/djt-labs/paaro/vms/components/djtx"
)

var _ pubsub.Filterer = &filterer{}

type filterer struct {
	tx *Tx
}

func NewPubSubFilterer(tx *Tx) pubsub.Filterer {
	return &filterer{tx: tx}
}

// Apply the filter on the addresses.
func (f *filterer) Filter(filters []pubsub.Filter) ([]bool, interface{}) {
	resp := make([]bool, len(filters))
	for _, utxo := range f.tx.UTXOs() {
		addressable, ok := utxo.Out.(djtx.Addressable)
		if !ok {
			continue
		}

		for _, address := range addressable.Addresses() {
			for i, c := range filters {
				if resp[i] {
					continue
				}
				resp[i] = c.Check(address)
			}
		}
	}
	return resp, api.JSONTxID{
		TxID: f.tx.ID(),
	}
}
