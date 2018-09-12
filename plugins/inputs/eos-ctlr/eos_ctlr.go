package eos_ctlr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aristanetworks/goeapi"
	"github.com/aristanetworks/goeapi/module"
	"github.com/mchuang3/telegraf"
	"github.com/mchuang3/telegraf/plugins/inputs"
)

const ethPrefix = "Ethernet"
const measurement = "port_stats"

type EosCtlrStats struct {
	connected   bool
	host        string
	client      *goeapi.Node
	MgmtAddress string `toml:"mgmt_address"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
}

func connectDB(s *EosCtlrStats) bool {
	var err error
	s.client, err = goeapi.Connect("http", s.MgmtAddress, s.Username, s.Password, goeapi.UseDefaultPortNum)
	if err != nil {
		return false
	}
	s.connected = true

	return true
}

func (s *EosCtlrStats) Disconnect() {
	// Nothing to do.
}

func (_ *EosCtlrStats) Description() string {
	return "Read metrics about Rack Controller managed Arista switch port statistics"
}

var sampleConfig = `
  ## Rack Controller Managed Arista statistics

  ## Switch Information
  mgmt_address = "10.0.0.5"
  username = "admin"
  password = "admin123"
`

func (_ *EosCtlrStats) SampleConfig() string {
	return sampleConfig
}

func (s *EosCtlrStats) Gather(acc telegraf.Accumulator) error {
	if !s.connected {
		if !connectDB(s) {
			return fmt.Errorf("Failed to connect to Arista Switch")
		}
	}

	now := time.Now()

	sh := module.Show(s.client)
	showData := sh.ShowInterfaces()
	for n, v := range showData.Interfaces {
		// Only care about Ethernet interfaces that are linked.
		if !(strings.HasPrefix(n, ethPrefix) && v.InterfaceStatus == "connected") {
			continue
		}

		// Parse port role from information in the Description,
		// which is something like the following:
		//   {"role":"local","svid":889,"conn_type":"none"}
		role := ""
		var roleInfo struct {
			Role string `json:"role"`
		}
		if err := json.Unmarshal([]byte(v.Description), &roleInfo); err == nil {
			role = roleInfo.Role
		}

		intfName := strings.TrimPrefix(n, ethPrefix)

		tags := map[string]string{
			"port": intfName,
			"role": role,
		}

		rxPkts := v.InterfaceCounters.InUcastPkts +
			v.InterfaceCounters.InMulticastPkts +
			v.InterfaceCounters.InBroadcastPkts
		txPkts := v.InterfaceCounters.OutUcastPkts +
			v.InterfaceCounters.OutMulticastPkts +
			v.InterfaceCounters.OutBroadcastPkts

		fields := map[string]interface{}{
			"rx_packets": float64(rxPkts),
			"rx_bytes":   float64(v.InterfaceCounters.InOctets),
			"rx_errors":  float64(v.InterfaceCounters.TotalInErrors),
			"tx_packets": float64(txPkts),
			"tx_bytes":   float64(v.InterfaceCounters.OutOctets),
			"tx_errors":  float64(v.InterfaceCounters.TotalOutErrors),
		}

		acc.AddCounter(measurement, fields, tags, now)
	}

	return nil
}

func init() {
	inputs.Add("eos-ctlr", func() telegraf.Input {
		return &EosCtlrStats{
			connected: false,
		}
	})
}
