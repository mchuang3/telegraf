package sim

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mchuang3/telegraf"
	"github.com/mchuang3/telegraf/plugins/inputs"
	gops "github.com/shirou/gopsutil/net"
)

const sysStats = "/proc/net/dev"
const tmpStats = "/tmp/stats.tmp"
const measurement = "port_stats"

type SimStats struct {
	StatsFile string
	Namespace string `toml:"namespace"`
}

func (_ *SimStats) Description() string {
	return "Read metrics about SimSwitch port statistics"
}

var sampleConfig = `
  ## SimSwitch statistics

  ## Switch Namespace
  namespace = "swns"
`

func (_ *SimStats) SampleConfig() string {
	return sampleConfig
}

func (s *SimStats) Gather(acc telegraf.Accumulator) error {
	// Get a copy of stats file where we can read it.
	var cmd *exec.Cmd
	if s.Namespace != "" {
		cmd = exec.Command("ip", "netns", "exec", s.Namespace, "/bin/cp", s.StatsFile, tmpStats)
	} else {
		cmd = exec.Command("/bin/cp", s.StatsFile, tmpStats)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Exec cp error: %v", err)
	}
	defer os.Remove(tmpStats)

	// Get counters.
	netio, err := gops.IOCountersByFile(true, tmpStats)
	if err != nil {
		return fmt.Errorf("Error extracting counters: %s", err)
	}

	now := time.Now()
	for _, io := range netio {
		// Skip loopback and any interface with "_" in its name.
		if io.Name == "lo" {
			continue
		}
		pieces := strings.Split(io.Name, "_")
		if len(pieces) > 1 {
			continue
		}

		tags := map[string]string{
			"port": io.Name,
		}

		fields := map[string]interface{}{
			"rx_packets": float64(io.PacketsRecv),
			"rx_bytes":   float64(io.BytesRecv),
			"tx_packets": float64(io.PacketsSent),
			"tx_bytes":   float64(io.BytesSent),
			"rx_errors":  float64(io.Errin),
			"tx_errors":  float64(io.Errout),
		}
		acc.AddCounter(measurement, fields, tags, now)
	}

	return nil
}

func init() {
	inputs.Add("sim", func() telegraf.Input {
		return &SimStats{
			StatsFile: sysStats,
		}
	})
}
