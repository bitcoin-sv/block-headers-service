package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type latestBlockMetrics struct {
	height    *prometheus.GaugeVec
	timestamp *prometheus.GaugeVec
}

func registerLatestBlockMetrics(reg prometheus.Registerer) *latestBlockMetrics {
	return &latestBlockMetrics{
		height:    registerGaugeVec(reg, latestBlockHeightName, []string{"state"}),
		timestamp: registerGaugeVec(reg, latestBlockTimestampName, []string{"state"}),
	}
}

func SetLatestBlock(height int32, timestamp time.Time, state string) {
	if metrics, enabled := Get(); enabled {
		metrics.latestBlock.height.WithLabelValues(state).Set(float64(height))
		metrics.latestBlock.timestamp.WithLabelValues(state).Set(float64(timestamp.Unix()))
	}
}
