package nx9k_ctlr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mchuang3/telegraf"
	"github.com/mchuang3/telegraf/plugins/inputs"
)

const measurement = "port_stats"

type Nx9kCtlrStats struct {
	connected   bool
	host        string
	client      *http.Client
	MgmtAddress string `toml:"mgmt_address"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
}

func connectDB(s *Nx9kCtlrStats) bool {
	s.host = "http://" + s.MgmtAddress + "/ins"
	s.client = &http.Client{}
	s.connected = true

	return true
}

func (s *Nx9kCtlrStats) Disconnect() {
	// Nothing to do.
}

func (_ *Nx9kCtlrStats) Description() string {
	return "Read metrics about Rack Controller managed Cisco NX9000 port statistics"
}

var sampleConfig = `
  ## Rack Controller Managed Cisco NX9000 statistics

  ## Switch Information
  mgmt_address = "10.0.0.5"
  username = "admin"
  password = "admin123"
`

func (_ *Nx9kCtlrStats) SampleConfig() string {
	return sampleConfig
}

type JsonRpcRequest struct {
	JSON_RPC string       `json:"jsonrpc"`
	Method   string       `json:"method"`
	Params   JsonRpcParam `json:"params"`
	Id       int          `json:"id"`
}

type JsonRpcParam struct {
	Cmd string `json:"cmd"`
	Ver int    `json:"version"`
}

type JsonRpcResponse struct {
	JSON_RPC string        `json:"jsonrpc"`
	Result   JsonResult    `json:"result"`
	Error    *JsonRpcError `json:"error,omitempty"`
	Id       int           `json:"id"`
}

type JsonResult struct {
	Body map[string]interface{} `json:"body"`
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const ethPrefix = "Ethernet1/"

func newJsonRpcRequest(cmds ...string) ([]byte, error) {
	var commands []JsonRpcRequest

	for i, cmd := range cmds {
		req := JsonRpcRequest{
			JSON_RPC: "2.0",
			Method:   "cli",
			Params: JsonRpcParam{
				Cmd: cmd,
				Ver: 1,
			},
			Id: i + 1, // 1-based
		}
		commands = append(commands, req)
	}

	out, err := json.Marshal(commands)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Nx9kCtlrStats) execCmd(cmds ...string) ([]JsonRpcResponse, error) {
	command, err := newJsonRpcRequest(cmds...)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.host, bytes.NewBuffer(command))
	if err != nil {
		log.Printf("E! HTTP New request failed - %v", err)
		return nil, err
	}

	req.Header.Set("content-type", "application/json-rpc")
	req.SetBasicAuth(s.Username, s.Password)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("E! HTTP request failed - %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("E! HTTP error - received status %v (%d) from server",
			resp.Status, resp.StatusCode)
		// Attempt to read & return the response body anyway
		// so caller can do what they want w/ error data.
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("E! HTTP response body read failed - %v", err)
		return nil, err
	}

	isSlice := false
	for _, c := range body {
		// http://godoc.org/encoding/json?file=scanner.go#isSpace
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		isSlice = c == '['
		break
	}

	var responses []JsonRpcResponse
	if isSlice {
		if err = json.Unmarshal(body, &responses); err != nil {
			return nil, err
		}
	} else {
		var resp JsonRpcResponse
		if err = json.Unmarshal(body, &resp); err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// Cisco responses for things that may have 1 or more items are returned in
// different JSON format (individual item or array).  This function normalizes
// the output so it's always in an array to simplify processing.
func normalizeSlice(r interface{}) []interface{} {
	rows := make([]interface{}, 0)
	switch r.(type) {
	case map[string]interface{}:
		rows = append(rows, r)
	case []interface{}:
		rows = r.([]interface{})
	default:
		log.Printf("E! Unexpected data type - %T", r)
	}
	return rows
}

func normalizeStats(stat interface{}) float64 {
	switch stat.(type) {
	case string:
		val, _ := strconv.Atoi(stat.(string))
		return float64(val)
	case float64:
		return stat.(float64)
	default:
		return float64(0)
	}
}

func (s *Nx9kCtlrStats) Gather(acc telegraf.Accumulator) error {
	if !s.connected {
		if !connectDB(s) {
			return fmt.Errorf("Failed to connect to Cisco NX9K Switch")
		}
	}

	now := time.Now()
	responses, err := s.execCmd("show interface")
	if err != nil || len(responses) < 1 {
		return fmt.Errorf("Failed to get interface data - %v", err)
	}

	if t, ok := responses[0].Result.Body["TABLE_interface"]; ok {
		if r, ok := t.(map[string]interface{})["ROW_interface"]; ok {
			row := normalizeSlice(r)
			for _, intf := range row {
				intfMap := intf.(map[string]interface{})

				// Skip mgmt0 interface
				name, ok := intfMap["interface"]
				if !ok || name.(string) == "mgmt0" {
					continue
				}
				intfName := strings.TrimPrefix(name.(string), ethPrefix)

				// Skip interfaces that are unlinked.
				linkState, ok := intfMap["state"]
				if !ok || linkState.(string) != "up" {
					continue
				}

				// Parse port role.
				role := ""
				if desc, ok := intfMap["desc"]; ok {
					// Output of "desc" is something like the following:
					//   {"role":"local","svid":889,"conn_type":"none"}
					var roleInfo struct {
						Role string `json:"role"`
					}
					if err := json.Unmarshal([]byte(desc.(string)), &roleInfo); err == nil {
						role = roleInfo.Role
					}
				}

				tags := map[string]string{
					"port": intfName,
					"role": role,
				}

				// NOTE: there's some inconsistency in Cisco stats output.
				// Some have ""s around them, making them string.  Others don't,
				// making them float64.  Make them all float64 to be consistent.
				// E.g.
				//    "eth_inucast": 79929,
				//    "eth_inbytes": 5152022,
				//    "eth_jumbo_inpkts": "0",
				//    "eth_storm_supp": "0",
				//    "eth_runts": 0,
				//    "eth_giants": 0,
				fields := map[string]interface{}{
					"rx_packets": normalizeStats(intfMap["eth_inpkts"]),
					"rx_bytes":   normalizeStats(intfMap["eth_inbytes"]),
					"rx_errors":  normalizeStats(intfMap["eth_inerr"]),
					"tx_packets": normalizeStats(intfMap["eth_outpkts"]),
					"tx_bytes":   normalizeStats(intfMap["eth_outbytes"]),
					"tx_errors":  normalizeStats(intfMap["eth_outerr"]),
				}

				acc.AddCounter(measurement, fields, tags, now)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("nx9k-ctlr", func() telegraf.Input {
		return &Nx9kCtlrStats{
			connected: false,
		}
	})
}
