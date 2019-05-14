package nx9k_ctlr

import (
	"math/rand"
	"time"

	"github.com/mchuang3/telegraf"
)

type stats struct {
	name       string
	role       string
	rx_packets uint64
	rx_bytes   uint64
	tx_packets uint64
	tx_bytes   uint64
	rx_errors  uint64
	tx_errors  uint64
}

var portList []string
var portStats map[string]*stats

func init() {
	// Server ports 1-20, Uplink 47, Rack Controller port 48.
	serverPorts := []string{"1", "2", "3", "4", "5", "6", "7",
		"8", "9", "10", "11", "12", "13", "14", "15",
		"16", "17", "18", "19", "20"}
	uplinkPorts := []string{"47"}
	rackCtlPorts := []string{"48"}

	// Initialize stats.
	portStats = make(map[string]*stats)
	for _, p := range serverPorts {
		portStats[p] = &stats{
			name: p,
			role: "server",
		}
	}
	for _, p := range uplinkPorts {
		portStats[p] = &stats{
			name: p,
			role: "uplink",
		}
	}
	for _, p := range rackCtlPorts {
		portStats[p] = &stats{
			name: p,
			role: "ctlr",
		}
	}
}

func gatherSimStats(acc telegraf.Accumulator) error {
	now := time.Now()
	for _, p := range portStats {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)

		p.rx_packets += uint64(r1.Intn(5000))
		p.rx_bytes += uint64(r1.Intn(1000) * 1024)
		p.tx_packets += uint64(r1.Intn(500))
		p.tx_bytes += uint64(r1.Intn(100) * 1024)
		p.rx_errors += uint64(r1.Intn(5))
		p.tx_errors += uint64(r1.Intn(2))

		tags := map[string]string{
			"port": p.name,
			"role": p.role,
		}

		fields := map[string]interface{}{
			"rx_packets": float64(p.rx_packets),
			"rx_bytes":   float64(p.rx_bytes),
			"tx_packets": float64(p.tx_packets),
			"tx_bytes":   float64(p.tx_bytes),
			"rx_errors":  float64(p.rx_errors),
			"tx_errors":  float64(p.tx_errors),
		}
		acc.AddCounter(measurement, fields, tags, now)
	}
	return nil
}
