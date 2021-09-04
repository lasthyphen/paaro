// (c) 2020 - 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package common

import (
	"github.com/djt-labs/paaro/ids"
	"github.com/djt-labs/paaro/snow"
	"github.com/djt-labs/paaro/snow/validators"
)

// DefaultConfigTest returns a test configuration
func DefaultConfigTest() Config {
	isBootstrapped := false
	subnet := &SubnetTest{
		IsBootstrappedF: func() bool { return isBootstrapped },
		BootstrappedF:   func(ids.ID) { isBootstrapped = true },
	}

	return Config{
		Ctx:                           snow.DefaultContextTest(),
		Validators:                    validators.NewSet(),
		Beacons:                       validators.NewSet(),
		Sender:                        &SenderTest{},
		Bootstrapable:                 &BootstrapableTest{},
		Subnet:                        subnet,
		Timer:                         &TimerTest{},
		MultiputMaxContainersSent:     2000,
		MultiputMaxContainersReceived: 2000,
	}
}
