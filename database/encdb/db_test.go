// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package encdb

import (
	"testing"

	"github.com/lasthyphen/paaro/database"
	"github.com/lasthyphen/paaro/database/memdb"
)

func TestInterface(t *testing.T) {
	pw := "lol totally a secure password"
	for _, test := range database.Tests {
		unencryptedDB := memdb.New()
		db, err := New([]byte(pw), unencryptedDB)
		if err != nil {
			t.Fatal(err)
		}

		test(t, db)
	}
}

func BenchmarkInterface(b *testing.B) {
	pw := "lol totally a secure password"
	for _, size := range database.BenchmarkSizes {
		keys, values := database.SetupBenchmark(b, size[0], size[1], size[2])
		for _, bench := range database.Benchmarks {
			unencryptedDB := memdb.New()
			db, err := New([]byte(pw), unencryptedDB)
			if err != nil {
				b.Fatal(err)
			}
			bench(b, db, "encdb", keys, values)
		}
	}
}
