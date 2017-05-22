# Rack Controller Managed Cisco NX9000 Physical Interface Statistics Plugin

#### Description

The Rack Controller managed Cisco NX9000 Telegraf plugin uses Cisco NX-OS API over http
to gather interface statistics from the switch.  Statistics are only gathered on
physical interfaces, i.e. interface type of "system".  Other interface types such
as "internal", "loopback", etc. are currently ignored.

### Configuration:

To enable Cisco NX9000 interface statistics collection, enable **inputs.nx9k-ctlr**
configuration in the **telegraf.conf** file.  In addition, the globally unique
switch_id should be set in the **[global_tags]** field in telegraf.conf:

```
 [global_tags]
   # dc = "us-east-1" # will tag all metrics with dc=us-east-1
   # rack = "1a"
   switch = "b1668a18-7659-47ac-91ea-bf7357f2dfad"
   ## Environment variables can be used as tags, and throughout the config file
   # user = "$USER"

   ...

 # Read metrics about Rack Controller managed Cisco NX9000 port statistics
 [[inputs.nx9k-ctlr]]
   ## Rack Controller Managed Cisco NX9000 statistics
 
   ## Switch Information
   mgmt_address = "10.0.0.5"
   username = "admin"
   password = "admin123"
```

### Measurements & Fields:

Currently, the following measurements are collected from each interface:

   - rx_packets
   - rx_bytes
   - rx_errors
   - tx_packets
   - tx_bytes
   - tx_errors

### Tags:

 - globally set **rack_id**, **switch_id**, etc.
 - interface **name**
 - interface **role** (e.g. "uplink", "local", "bmc", etc.)

### Example Output:

```
% ./telegraf -config ./telegraf.conf -test
* Plugin: nx0k-ctlr, Collection 1
> ops,host=switch,port=47,role=uplink,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=26,role=local,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=13,role=local,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
> ops,host=switch,port=8,role=local,switch_id=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=0,rx_errors=0,rx_packets=0,tx_bytes=0,tx_errors=0,tx_packets=0 1474300899000000000
...
```
