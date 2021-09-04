package metercacher

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/djt-labs/paaro/cache"
)

func TestInterface(t *testing.T) {
	for _, test := range cache.CacherTests {
		cache := &cache.LRU{Size: test.Size}
		c, err := New("", prometheus.NewRegistry(), cache)
		if err != nil {
			t.Fatal(err)
		}

		test.Func(t, c)
	}
}
