# OpenSwitch Physical Interface Statistics Plugin

#### Description

The OpenSwitch Telegraf plugin uses the libovsdb API to gather interface statistics
from the OVSDB.  Statistics are only gathered on physical interfaces, i.e. interface
type of "system".  Other interface types such as "internal", "loopback", etc. are
currently ignored.

### Configuration:

To enable OpenSwitch interface statistics collection, enable **inputs.ops** configuration
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

 # Read metrics about OpenSwitch system & port statistics
 [[inputs.ops]]
   ## OpenSwitch statistics
 
   ## OVSDB socket connection endpoint
   ovsdb_socket = "/var/run/openswitch/db.sock"
````

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

 - globally set **switch_id**
 - interface **name**

### Example Output:

```
% ./telegraf -config ./telegraf.conf -test
* Plugin: ops, Collection 1
> ops,host=switch,port=47,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=26,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=13,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=8,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad ipv4_mc_rx_packets=0,ipv4_mc_tx_packets=0,ipv4_uc_rx_packets=0,ipv4_uc_tx_packets=0,ipv6_mc_rx_packets=0,ipv6_mc_tx_packets=0,ipv6_uc_rx_packets=0,ipv6_uc_tx_packets=0,rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
...
```
