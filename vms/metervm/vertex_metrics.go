// (c) 2021, Dijets, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/djt-labs/paaro/utils/metric"
	"github.com/djt-labs/paaro/utils/wrappers"
)

type vertexMetrics struct {
	pending,
	parse,
	get metric.Averager
}

func (m *vertexMetrics) Initialize(
	namespace string,
	reg prometheus.Registerer,
) error {
	errs := wrappers.Errs{}
	m.pending = newAverager(namespace, "pending_txs", reg, &errs)
	m.parse = newAverager(namespace, "parse_tx", reg, &errs)
	m.get = newAverager(namespace, "get_tx", reg, &errs)
	return errs.Err
}
