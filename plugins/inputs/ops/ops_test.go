package ops

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/cenk/rpc2"
	"github.com/cenk/rpc2/jsonrpc"
	db "github.com/dancripe/libovsdb"
	"github.com/mchuang3/telegraf/testutil"
	uuidPkg "github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//-------------------------------------
// OpenSwitch OVSDB Server Simulator
//-------------------------------------
const testOvsSocket = "testDB.sock"

type testTable map[string]map[string]interface{}

var simInterfaceTbl testTable

// Copied from libovsdb package, file notation.go.
type OpResult struct {
	Count   int                      `json:"count,omitempty"`
	Error   string                   `json:"error,omitempty"`
	Details string                   `json:"details,omitempty"`
	UUID    db.UUID                  `json:"uuid,omitempty"`
	Rows    []map[string]interface{} `json:"rows,omitempty"`
}

func ovsdbListDB(client *rpc2.Client, params []interface{}, reply *interface{}) error {
	var rep [1]string
	rep[0] = "OpenSwitch"
	*reply = rep
	return nil
}

func ovsdbGetSchema(client *rpc2.Client, params []interface{}, reply *interface{}) error {
	if params[0].(string) != "OpenSwitch" {
		return fmt.Errorf("unknown database")
	}

	var iface db.TableSchema
	iface.Columns = make(map[string]db.ColumnSchema)
	iface.Columns["name"] = db.ColumnSchema{"", "string", false, false}
	iface.Columns["type"] = db.ColumnSchema{"", "string", false, true}
	iface.Columns["link_state"] = db.ColumnSchema{"", "string", false, true}
	iface.Columns["statistics"] = db.ColumnSchema{"", "map", false, true}
	iface.Columns["external_ids"] = db.ColumnSchema{"", "map", false, true}

	var resp db.DatabaseSchema
	resp.Tables = make(map[string]db.TableSchema)
	resp.Name = "OpenSwitch"
	resp.Version = "0.1.8"
	resp.Tables["Interface"] = iface

	*reply = &resp

	return nil
}

func opIsSelectAll(oper map[string]interface{}) bool {
	if oper["op"] != "select" {
		return false
	}

	where := oper["where"].([]interface{})
	cond := where[0].([]interface{})
	if len(cond) != 3 || cond[0] != "_uuid" || cond[1] != "!=" {
		return false
	}
	uuidArray := cond[2].([]interface{})
	if len(uuidArray) != 2 || uuidArray[0] != "uuid" ||
		uuidArray[1] != "00000000-0000-0000-0000-000000000000" {
		return false
	}

	return true
}

func handleInterfaceTable(params *[]interface{}, oper map[string]interface{}, results *[]OpResult) bool {
	var result OpResult

	if opIsSelectAll(oper) {
		result.Rows = []map[string]interface{}{}
		for _, intf := range simInterfaceTbl {
			result.Rows = append(result.Rows, intf)
		}
		*results = append(*results, result)
		return true
	}

	result.Error = "unsupported"
	result.Details = "unsupported operation for Interface table: " + oper["op"].(string)
	*results = append(*results, result)
	return false
}

func ovsdbTransact(client *rpc2.Client, params []interface{}, reply *interface{}) error {
	if len(params) < 2 {
		*reply = nil
		return fmt.Errorf("bad param")
	}

	if params[0].(string) != "OpenSwitch" {
		*reply = nil
		return fmt.Errorf("unknown database")
	}

	var results = []OpResult{}
	params = params[1:]

	for len(params) > 0 {
		oper := params[0].(map[string]interface{})
		params = params[1:]

		switch oper["table"] {
		case "Interface":
			if handleInterfaceTable(&params, oper, &results) {
				break
			}

		default:
			var result OpResult
			result.Error = "unsupported"
			result.Details = "unsupported table " + oper["table"].(string)
			results = append(results, result)
			break
		}
	}

	*reply = results

	return nil
}

var stats_1 = map[string]interface{}{
	"rx_packets":         float64(39492302),
	"rx_bytes":           float64(349),
	"rx_errors":          float64(55),
	"tx_packets":         float64(394922),
	"tx_bytes":           float64(3493),
	"tx_errors":          float64(55),
	"ipv4_uc_rx_packets": float64(33994),
	"ipv4_uc_tx_packets": float64(2123),
	"ipv4_mc_rx_packets": float64(7833),
	"ipv4_mc_tx_packets": float64(2335894),
	"ipv6_uc_rx_packets": float64(3453452235235),
	"ipv6_uc_tx_packets": float64(94983),
	"ipv6_mc_rx_packets": float64(68826623),
	"ipv6_mc_tx_packets": float64(338682),
}

var stats_2 = map[string]interface{}{
	"rx_packets": float64(123),
	"rx_bytes":   float64(0),
}

var stats_24_3 = map[string]interface{}{
	"rx_packets":         float64(128467827638694),
	"rx_bytes":           float64(34384389),
	"rx_errors":          float64(58333225),
	"tx_packets":         float64(394922383817162003848),
	"tx_bytes":           float64(201),
	"tx_errors":          float64(333),
	"ipv4_uc_rx_packets": float64(127994),
	"ipv4_uc_tx_packets": float64(212),
	"ipv4_mc_rx_packets": float64(7833),
	"ipv4_mc_tx_packets": float64(2335894),
	"ipv6_uc_rx_packets": float64(3453452235235),
	"ipv6_uc_tx_packets": float64(94983),
	"ipv6_mc_rx_packets": float64(68826623),
	"ipv6_mc_tx_packets": float64(338682),
}

var stats_bridge_normal = map[string]interface{}{
	"rx_packets": float64(50),
	"rx_bytes":   float64(500),
}

var stats_lo = map[string]interface{}{
	"rx_packets": float64(100),
	"rx_bytes":   float64(1000),
}

func setupDefaultTables() {
	simInterfaceTbl = make(testTable)

	row := make(map[string]interface{})
	row["_uuid"] = db.UUID{GoUuid: uuidPkg.New()}
	row["name"] = "1"
	row["type"] = "system"
	row["link_state"] = "up"
	statsMap, _ := db.NewOvsMap(stats_1)
	row["statistics"] = statsMap
	extIds := make(map[string]string)
	extIds["role"] = "local"
	extIdsMap, _ := db.NewOvsMap(extIds)
	row["external_ids"] = *extIdsMap
	simInterfaceTbl["1"] = row

	row = make(map[string]interface{})
	row["_uuid"] = db.UUID{GoUuid: uuidPkg.New()}
	row["name"] = "2"
	row["type"] = "system"
	row["link_state"] = "down"
	statsMap, _ = db.NewOvsMap(stats_2)
	row["statistics"] = statsMap
	extIds["role"] = "local"
	extIdsMap, _ = db.NewOvsMap(extIds)
	row["external_ids"] = *extIdsMap
	simInterfaceTbl["2"] = row

	row = make(map[string]interface{})
	row["_uuid"] = db.UUID{GoUuid: uuidPkg.New()}
	row["name"] = "24-3"
	row["type"] = "system"
	row["link_state"] = "up"
	statsMap, _ = db.NewOvsMap(stats_24_3)
	row["statistics"] = statsMap
	extIds["role"] = "uplink"
	extIdsMap, _ = db.NewOvsMap(extIds)
	row["external_ids"] = *extIdsMap
	simInterfaceTbl["24-3"] = row

	row = make(map[string]interface{})
	row["_uuid"] = db.UUID{GoUuid: uuidPkg.New()}
	row["name"] = "bridge_normal"
	row["type"] = "internal"
	row["link_state"] = "up"
	statsMap, _ = db.NewOvsMap(stats_bridge_normal)
	row["statistics"] = statsMap
	simInterfaceTbl["bridge_normal"] = row

	row = make(map[string]interface{})
	row["_uuid"] = db.UUID{GoUuid: uuidPkg.New()}
	row["name"] = "lo"
	row["type"] = "loopback"
	row["link_state"] = "up"
	statsMap, _ = db.NewOvsMap(stats_lo)
	row["statistics"] = statsMap
	simInterfaceTbl["lo"] = row
}

func simOvsdbServer(wg *sync.WaitGroup, termCh chan bool) {
	log.Printf("[SIM-OVSDB] Starting server...\n")
	defer wg.Done()

	cmd := exec.Command("rm", "-f", testOvsSocket)
	cmd.Run()

	setupDefaultTables()

	// Set up JSON RPC server.
	srv := rpc2.NewServer()
	srv.Handle("list_dbs", ovsdbListDB)
	srv.Handle("get_schema", ovsdbGetSchema)
	srv.Handle("transact", ovsdbTransact)

	unixaddr, err := net.ResolveUnixAddr("unix", testOvsSocket)
	if err != nil {
		log.Printf("[SIM-OVSDB] Failed to resolve addr %s, %v\n", testOvsSocket, err)
		return
	}
	lis, err := net.ListenUnix("unix", unixaddr)
	if err != nil {
		log.Printf("[SIM-OVSDB] Failed to listen on socket unix:%s, %v\n",
			testOvsSocket, err)
		return
	}
	defer lis.Close()

LOOP:
	for {
		select {
		case <-termCh:
			log.Printf("[SIM-OVSDB] terminate signal received.  Exiting.\n")
			break LOOP
		default:
			lis.SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := lis.Accept()
			if err != nil {
				if err, ok := err.(net.Error); ok && !err.Timeout() {
					log.Printf("[SIM-OVSDB] rpc.serve: accept err=%v\n", err)
				}
			} else {
				log.Printf("[SIM-OVSDB] Accepted connection.")
				srv.ServeCodec(jsonrpc.NewJSONCodec(conn))
			}
		}
	}
}

//-------------------------------------------------------------------------

func TestOpsGatherStats(t *testing.T) {
	// Start simulated OVSDB server.
	var wg sync.WaitGroup
	termCh := make(chan bool, 1)
	wg.Add(1)
	go simOvsdbServer(&wg, termCh)
	time.Sleep(time.Second)
	defer func() {
		termCh <- true
		wg.Wait()
	}()

	// Telegraf OPS plugin
	o := OpsStats{
		OvsdbSocket: testOvsSocket,
	}

	var acc testutil.Accumulator
	err := o.Gather(&acc)

	require.NoError(t, err)

	// There should only be two data points since
	// port 2 is link down, and bridge_normal
	// and lo interfaces are ignored.
	assert.Equal(t, acc.NMetrics(), uint64(2))

	acc.AssertContainsTaggedFields(t, "port_stats", stats_1,
		map[string]string{"port": "1", "role": "local"})
	acc.AssertContainsTaggedFields(t, "port_stats", stats_24_3,
		map[string]string{"port": "24-3", "role": "uplink"})

	o.Disconnect()
}
