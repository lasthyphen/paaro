// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/lasthyphen/paaro/utils/metric"
	"github.com/lasthyphen/paaro/utils/wrappers"
)

type blockMetrics struct {
	buildBlock,
	parseBlock,
	getBlock,
	setPreference,
	lastAccepted metric.Averager
}

func (m *blockMetrics) Initialize(
	namespace string,
	reg prometheus.Registerer,
) error {
	errs := wrappers.Errs{}
	m.buildBlock = newAverager(namespace, "build_block", reg, &errs)
	m.parseBlock = newAverager(namespace, "parse_block", reg, &errs)
	m.getBlock = newAverager(namespace, "get_block", reg, &errs)
	m.setPreference = newAverager(namespace, "set_preference", reg, &errs)
	m.lastAccepted = newAverager(namespace, "last_accepted", reg, &errs)
	return errs.Err
}
