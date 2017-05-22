# Rack Controller Managed OpenSwitch Physical Interface Statistics Plugin

#### Description

The Rack Controller managed OpenSwitch Telegraf plugin uses ovs-vsctl tool over ssh
to gather interface statistics from the OVSDB.  Statistics are only gathered on
physical interfaces, i.e. interface type of "system", whose link state is up.
Other interface types such as "internal", "loopback", etc. or interfaces whose
links are down are currently ignored.

### Configuration:

To enable OpenSwitch interface statistics collection, enable **inputs.ops-ctlr** configuration
in the **telegraf.conf** file.  In addition, the globally unique switch_id should be set
in the **[global_tags]** field in telegraf.conf:

```
 [global_tags]
   # dc = "us-east-1" # will tag all metrics with dc=us-east-1
   # rack = "1a"
   switch = "b1668a18-7659-47ac-91ea-bf7357f2dfad"
   ## Environment variables can be used as tags, and throughout the config file
   # user = "$USER"

   ...

 # Read metrics about Rack Controller managed OpenSwitch port statistics
 [[inputs.ops-ctlr]]
   ## Rack Controller Managed OpenSwitch statistics
 
   ## Switch Information
   mgmt_address = "10.0.0.5"
   username = "root"
   password = ""
```

### Measurements & Fields:

Currently, the following measurements are collected from each interface:

   - rx_packets
   - rx_bytes
   - rx_errors
   - tx_packets
   - tx_bytes
   - tx_errors
   - ipv4_uc_rx_packets
   - ipv4_uc_tx_packets
   - ipv4_mc_rx_packets
   - ipv4_mc_tx_packets
   - ipv6_uc_rx_packets
   - ipv6_uc_tx_packets
   - ipv6_mc_rx_packets
   - ipv6_mc_tx_packets

### Tags:

 - globally set **rack_id**, **switch_id**, etc.
 - interface **name**
 - interface **role** (e.g. "uplink", "local", "bmc", etc.)

### Example Output:

```
% ./telegraf -config ./telegraf.conf -test
* Plugin: ops-ctlr, Collection 1
> ops,host=switch,port=47,role=uplink,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=26,role=local,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=13,role=local,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=8,role=local,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
...
```
