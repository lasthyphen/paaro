// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lasthyphen/paaro/utils/metric"
	"github.com/lasthyphen/paaro/utils/wrappers"
)

func newAverager(namespace, name string, reg prometheus.Registerer, errs *wrappers.Errs) metric.Averager {
	return metric.NewAveragerWithErrs(
		namespace,
		name,
		fmt.Sprintf("time (in ns) of a %s", name),
		reg,
		errs,
	)
}
