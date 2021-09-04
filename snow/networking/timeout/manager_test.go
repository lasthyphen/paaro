// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timeout

import (
	"sync"
	"testing"
	"time"

	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow/networking/benchlist"
	"github.com/djt-labs/paaro/utils/constants"
	"github.com/djt-labs/paaro/utils/timer"
	"github.com/prometheus/client_golang/prometheus"
)

func TestManagerFire(t *testing.T) {
	manager := Manager{}
	benchlist := benchlist.NewNoBenchlist()
	err := manager.Initialize(
		&timer.AdaptiveTimeoutConfig{
			InitialTimeout:     time.Millisecond,
			MinimumTimeout:     time.Millisecond,
			MaximumTimeout:     10 * time.Second,
			TimeoutCoefficient: 1.25,
			TimeoutHalflife:    5 * time.Minute,
		},
		benchlist,
		"",
		prometheus.NewRegistry(),
	)
	if err != nil {
		t.Fatal(err)
	}
	go manager.Dispatch()

	wg := sync.WaitGroup{}
	wg.Add(1)

	manager.RegisterRequest(ids.ShortID{}, ids.ID{}, constants.PullQueryMsg, ids.GenerateTestID(), wg.Done)

	wg.Wait()
}

func TestManagerCancel(t *testing.T) {
	manager := Manager{}
	benchlist := benchlist.NewNoBenchlist()
	err := manager.Initialize(
		&timer.AdaptiveTimeoutConfig{
			InitialTimeout:     time.Millisecond,
			MinimumTimeout:     time.Millisecond,
			MaximumTimeout:     10 * time.Second,
			TimeoutCoefficient: 1.25,
			TimeoutHalflife:    5 * time.Minute,
		},
		benchlist,
		"",
		prometheus.NewRegistry(),
	)
	if err != nil {
		t.Fatal(err)
	}
	go manager.Dispatch()

	wg := sync.WaitGroup{}
	wg.Add(1)

	fired := new(bool)

	id := ids.GenerateTestID()
	manager.RegisterRequest(ids.ShortID{}, ids.ID{}, constants.PullQueryMsg, id, func() { *fired = true })

	manager.RegisterResponse(ids.ShortID{}, ids.ID{}, id, constants.GetMsg, 1*time.Second)

	manager.RegisterRequest(ids.ShortID{}, ids.ID{}, constants.PullQueryMsg, ids.GenerateTestID(), wg.Done)

	wg.Wait()

	if *fired {
		t.Fatalf("Should have cancelled the function")
	}
}
