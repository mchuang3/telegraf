package ops

import (
	"fmt"
	"log"
	"time"

	db "github.com/dancripe/libovsdb"
	"github.com/mchuang3/telegraf"
	"github.com/mchuang3/telegraf/plugins/inputs"
)

const interfaceTable = "Interface"
const defaultOvsdbSocket = "/var/run/openvswitch/db.sock"

const measurement = "port_stats"

type OpsStats struct {
	connected bool
	client    *db.OvsdbClient

	OvsdbSocket string `toml:"ovsdb_socket"`
}

func connectDB(s *OpsStats) bool {
	handle, err := db.Dial("unix", s.OvsdbSocket)
	if err != nil {
		log.Printf("telegraf OPS plugin OVSDB client dial failed: %v", err)
		return false
	}
	s.client = handle
	s.connected = true
	return true
}

func dbTransact(ovs *db.OvsdbClient, ops ...db.Operation) ([]db.OperationResult, error) {
	reply, err := ovs.Transact("OpenSwitch", ops...)
	if err != nil {
		return reply, err
	}

	for i, o := range reply {
		if o.Error != "" {
			if i < len(ops) {
				log.Printf("dbTransact Ops %v failed: %v", i, ops[i])
			}
			return reply, fmt.Errorf("Transaction failed: %v (%v)", o.Error, o.Details)
		}
	}

	return reply, err
}

func (s *OpsStats) Disconnect() {
	if s.connected {
		s.client.Disconnect()
	}
}

func (_ *OpsStats) Description() string {
	return "Read metrics about OpenSwitch port statistics"
}

var sampleConfig = `
  ## OpenSwitch statistics

  ## OVSDB socket connection endpoint.
  ovsdb_socket = "/var/run/openvswitch/db.sock"
`

func (_ *OpsStats) SampleConfig() string {
	return sampleConfig
}

func (s *OpsStats) Gather(acc telegraf.Accumulator) error {
	if !s.connected {
		if !connectDB(s) {
			return fmt.Errorf("Failed to connect to OVSDB")
		}
	}

	interfaceColumns := []string{
		"_uuid",
		"name",
		"type",
		"link_state",
		"statistics",
	}
	cond := db.NewCondition("_uuid", "!=", db.UUID{GoUuid: "00000000-0000-0000-0000-000000000000"})
	op := db.Operation{
		Op:      "select",
		Table:   interfaceTable,
		Columns: interfaceColumns,
		Where:   []interface{}{cond},
	}

	reply, err := dbTransact(s.client, op)
	if err != nil {
		return fmt.Errorf("Failed OVSDB client transact for Interface stats: %v", err)
	}

	now := time.Now()
	for _, row := range reply[0].Rows {
		// Skip non-'system' interfaces, e.g. loopback, internal, etc.
		if ifType, ok := row.Fields["type"]; ok {
			if ifType.(string) != "system" {
				continue
			}
		}
		// Skip interfaces that are unlinked.
		if link, ok := row.Fields["link_state"]; ok {
			if link.(string) != "up" {
				continue
			}
		}

		ifName := row.Fields["name"].(string)
		stats, ok := row.Fields["statistics"]
		if !ok {
			log.Printf("Interface %v statistics missing", ifName)
			continue
		}

		statsMap := stats.(db.OvsMap)

		tags := map[string]string{
			"port": ifName,
		}

		fields := map[string]interface{}{
			"rx_packets":         statsMap.GoMap["rx_packets"],
			"rx_bytes":           statsMap.GoMap["rx_bytes"],
			"rx_errors":          statsMap.GoMap["rx_errors"],
			"tx_packets":         statsMap.GoMap["tx_packets"],
			"tx_bytes":           statsMap.GoMap["tx_bytes"],
			"tx_errors":          statsMap.GoMap["tx_errors"],
			"ipv4_uc_rx_packets": statsMap.GoMap["ipv4_uc_rx_packets"],
			"ipv4_uc_tx_packets": statsMap.GoMap["ipv4_uc_tx_packets"],
			"ipv4_mc_rx_packets": statsMap.GoMap["ipv4_mc_rx_packets"],
			"ipv4_mc_tx_packets": statsMap.GoMap["ipv4_mc_tx_packets"],
			"ipv6_uc_rx_packets": statsMap.GoMap["ipv6_uc_rx_packets"],
			"ipv6_uc_tx_packets": statsMap.GoMap["ipv6_uc_tx_packets"],
			"ipv6_mc_rx_packets": statsMap.GoMap["ipv6_mc_rx_packets"],
			"ipv6_mc_tx_packets": statsMap.GoMap["ipv6_mc_tx_packets"],
		}

		acc.AddCounter(measurement, fields, tags, now)
	}

	return nil
}

func init() {
	inputs.Add("ops", func() telegraf.Input {
		return &OpsStats{
			connected:   false,
			OvsdbSocket: defaultOvsdbSocket,
		}
	})
}
