package platformvm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/lasthyphen/paaro/ids"

	"github.com/lasthyphen/paaro/vms/components/djtx"
)

func TestBaseTxMarshalJSON(t *testing.T) {
	vm, _ := defaultVM()
	vm.ctx.Lock.Lock()
	defer func() {
		if err := vm.Shutdown(); err != nil {
			t.Fatal(err)
		}
		vm.ctx.Lock.Unlock()
	}()

	blockchainID := ids.ID{1}
	utxoTxID := ids.ID{2}
	assetID := ids.ID{3}
	fxID := ids.ID{4}
	tx := &BaseTx{BaseTx: djtx.BaseTx{
		BlockchainID: blockchainID,
		NetworkID:    4,
		Ins: []*djtx.TransferableInput{
			{
				FxID:   fxID,
				UTXOID: djtx.UTXOID{TxID: utxoTxID, OutputIndex: 5},
				Asset:  djtx.Asset{ID: assetID},
				In:     &djtx.TestTransferable{Val: 100},
			},
		},
		Outs: []*djtx.TransferableOutput{
			{
				FxID:  fxID,
				Asset: djtx.Asset{ID: assetID},
				Out:   &djtx.TestTransferable{Val: 100},
			},
		},
		Memo: []byte{1, 2, 3},
	}}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}
	asString := string(txBytes)
	switch {
	case !strings.Contains(asString, `"networkID":4`):
		t.Fatal("should have network ID")
	case !strings.Contains(asString, `"blockchainID":"SYXsAycDPUu4z2ZksJD5fh5nTDcH3vCFHnpcVye5XuJ2jArg"`):
		t.Fatal("should have blockchainID ID")
	case !strings.Contains(asString, `"inputs":[{"txID":"t64jLxDRmxo8y48WjbRALPAZuSDZ6qPVaaeDzxHA4oSojhLt","outputIndex":5,"assetID":"2KdbbWvpeAShCx5hGbtdF15FMMepq9kajsNTqVvvEbhiCRSxU","fxID":"2mB8TguRrYvbGw7G2UBqKfmL8osS7CfmzAAHSzuZK8bwpRKdY","input":{"Err":null,"Val":100}}]`):
		t.Fatal("inputs are wrong")
	case !strings.Contains(asString, `"outputs":[{"assetID":"2KdbbWvpeAShCx5hGbtdF15FMMepq9kajsNTqVvvEbhiCRSxU","fxID":"2mB8TguRrYvbGw7G2UBqKfmL8osS7CfmzAAHSzuZK8bwpRKdY","output":{"Err":null,"Val":100}}]`):
		t.Fatal("outputs are wrong")
	}
}
