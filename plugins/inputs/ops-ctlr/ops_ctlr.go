package ops

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	db "github.com/dancripe/libovsdb"
	"github.com/mchuang3/telegraf"
	"github.com/mchuang3/telegraf/plugins/inputs"
)

const measurement = "port_stats"

type OpsCtlrStats struct {
	connected   bool
	client      *ssh.Client
	MgmtAddress string `toml:"mgmt_address"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
}

func connectDB(s *OpsCtlrStats) bool {
	var config = &ssh.ClientConfig{
		User: s.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", s.MgmtAddress+":22", config)
	if err != nil {
		log.Printf("E! telegraf OPS-CTLR plugin client dial failed: %v", err)
		return false
	}

	s.client = conn
	s.connected = true

	return true
}

func (s *OpsCtlrStats) Disconnect() {
	if s.connected {
		s.client.Close()
	}
}

func (_ *OpsCtlrStats) Description() string {
	return "Read metrics about Rack Controller managed OpenSwitch port statistics"
}

var sampleConfig = `
  ## Rack Controller Managed OpenSwitch statistics

  ## Switch Information
  mgmt_address = "10.0.0.5"
  username = "root"
  password = ""
`

func (_ *OpsCtlrStats) SampleConfig() string {
	return sampleConfig
}

func getPortRole(c *ssh.Client, port string) string {
	session, err := c.NewSession()
	if err != nil {
		log.Printf("E! Failed to get new session: %v", err)
		return ""
	}
	defer session.Close()

	command := fmt.Sprintf("ovs-vsctl get Port %v external_ids", port)
	cmdOut, err := session.Output(command)
	if err != nil {
		log.Printf("E! Failed to get port role - %v", err)
		return ""
	}

	// Output of the command above is something like the following:
	//   {description="{\"role\":\"local\",\"svid\":3997,\"conn_type\":\"none\"}"}\n
	// Need to get rid of the \ and prefix/suffix, leaving
	//   {"role":"local","svid":3997,"conn_type":"none"}
	description := strings.TrimPrefix(string(cmdOut), `{description="`)
	description = strings.TrimSuffix(description, "\"}\n")
	description = strings.Replace(description, `\`, "", -1)

	var roleInfo struct {
		Role string `json:"role"`
	}

	if err := json.Unmarshal([]byte(description), &roleInfo); err == nil {
		return roleInfo.Role
	} else {
		return ""
	}
}

func (s *OpsCtlrStats) Gather(acc telegraf.Accumulator) error {
	if !s.connected {
		if !connectDB(s) {
			return fmt.Errorf("Failed to connect to OpenSwitch")
		}
	}

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to get new session: %v", err)
	}
	defer session.Close()

	command := `ovsdb-client transact '["OpenSwitch",{"op":"select","table":"Interface","where":[],"columns":["name", "type", "link_state", "statistics"]}]'`

	now := time.Now()
	cmdOut, err := session.Output(command)
	if err != nil {
		return fmt.Errorf("Failed to get port stats - %v", err)
	}

	var data []map[string][]map[string]interface{}
	if err := json.Unmarshal(cmdOut, &data); err != nil {
		return fmt.Errorf("Port stats unmarshal err - %v", err)
	}

	for _, entry := range data {
		if rows, ok := entry["rows"]; ok {
			for _, row := range rows {
				// Skip non-'system' interfaces, e.g. loopback, internal, etc.
				if ifType, ok := row["type"]; ok {
					if ifType.(string) != "system" {
						continue
					}
				}
				// Skip interfaces that are unlinked.
				if link, ok := row["link_state"]; ok {
					if link.(string) != "up" {
						continue
					}
				}

				ifName := row["name"].(string)
				tags := map[string]string{
					"port": ifName,
					"role": getPortRole(s.client, ifName),
				}

				stats, ok := row["statistics"]
				if !ok {
					log.Printf("W! Interface %v statistics missing", ifName)
					continue
				}

				var statsMap db.OvsMap

				if bs, err := json.Marshal(stats.([]interface{})); err == nil {
					if err = statsMap.UnmarshalJSON(bs); err == nil {
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
					} else {
						log.Printf("E! Failed to unmarshal stats - %v", err)
					}
				} else {
					log.Printf("E! Failed to convert stats into byte slice - %v", err)
				}
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("ops-ctlr", func() telegraf.Input {
		return &OpsCtlrStats{
			connected: false,
		}
	})
}
