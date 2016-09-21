package sim

import (
	"testing"

	"github.com/mchuang3/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var stats_1 = map[string]interface{}{
	"rx_packets": float64(12286),
	"rx_bytes":   float64(5599101),
	"rx_errors":  float64(0),
	"tx_packets": float64(12061),
	"tx_bytes":   float64(5912235),
	"tx_errors":  float64(834),
}

var stats_24_3 = map[string]interface{}{
	"rx_packets": float64(27163),
	"rx_bytes":   float64(94792615),
	"rx_errors":  float64(1),
	"tx_packets": float64(19323),
	"tx_bytes":   float64(14992843),
	"tx_errors":  float64(55),
}

func TestSimGatherStats(t *testing.T) {
	// Telegraf Sim plugin
	o := SimStats{
		StatsFile: "./testStats",
		Namespace: "",
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)

	require.NoError(t, err)

	assert.Equal(t, acc.NMetrics(), uint64(2))

	acc.AssertContainsTaggedFields(t, "port_stats", stats_1,
		map[string]string{"port": "1"})
	acc.AssertContainsTaggedFields(t, "port_stats", stats_24_3,
		map[string]string{"port": "24-3"})
}
