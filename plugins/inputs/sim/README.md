# SimSwitch Interface Statistics Plugin

#### Description

The SimSwitch Telegraf plugin gathers interface statistics from the simulated
switch ports running in "swns" namespace.  Statistics are only gathered on individual
ports.  Other interfaces such as VLAN, bond, bridges, etc. are currently ignored.

### Configuration:

To enable SimSwitch interface statistics collection, enable **inputs.sim** configuration
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

 # Read metrics about SimSwitch system & port statistics
 [[inputs.sim]]
   ## SimSwitch statistics

   ## Switch Namespace
   namespace = "swns"
```

### Measurements & Fields:

Currently, the following measurements are collected from each port:

   - rx_packets
   - rx_bytes
   - rx_errors
   - tx_packets
   - tx_bytes
   - tx_errors

### Tags:

 - globally set **switch_id**
 - port **name**

### Example Output:

```
openswitch@Switch-VM:~/Code/go/bin$ sudo ./telegraf -config telegraf.conf -test
* Plugin: cpu, Collection 1
* Plugin: cpu, Collection 2
> cpu,cpu=cpu0,host=Switch-VM,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad usage_guest=0,usage_guest_nice=0,usage_idle=99.99999999696836,usage_iowait=0,usage_irq=0,usage_nice=0,usage_softirq=0,usage_steal=0,usage_system=0,usage_user=0 1474485262000000000
> cpu,cpu=cpu1,host=Switch-VM,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad usage_guest=0,usage_guest_nice=0,usage_idle=95.99999999918509,usage_iowait=0,usage_irq=0,usage_nice=0,usage_softirq=0,usage_steal=0,usage_system=1.999999999998181,usage_user=1.999999999998181 1474485262000000000
> cpu,cpu=cpu-total,host=Switch-VM,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad usage_guest=0,usage_guest_nice=0,usage_idle=97.95918367165116,usage_iowait=0,usage_irq=0,usage_nice=0,usage_softirq=0,usage_steal=0,usage_system=1.0204081632231647,usage_user=1.0204081631999635 1474485262000000000
* Plugin: mem, Collection 1
> mem,host=Switch-VM,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad active=558559232i,available=729554944i,available_percent=35.952484280855444,buffered=13660160i,cached=557948928i,free=279732224i,inactive=530964480i,total=2029219840i,used=1299664896i,used_percent=64.04751571914456 1474485262000000000
* Plugin: system, Collection 1
> system,host=Switch-VM,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad load1=0,load15=0.17,load5=0.1,n_cpus=2i,n_users=8i 1474485262000000000
> system,host=Switch-VM,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad uptime=134619i,uptime_format="1 day, 13:23" 1474485262000000000
* Plugin: sim, Collection 1
>>> Processing stats for 1
> port_stats,host=Switch-VM,port=1,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=10782709,rx_errors=0,rx_packets=179604,tx_bytes=266135265,tx_errors=0,tx_packets=206076 1474485262000000000
>>> Processing stats for 2
> port_stats,host=Switch-VM,port=2,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=10782539,rx_errors=0,rx_packets=179610,tx_bytes=266118688,tx_errors=0,tx_packets=205987 1474485262000000000
>>> Processing stats for 3
> port_stats,host=Switch-VM,port=3,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=10782339,rx_errors=0,rx_packets=179606,tx_bytes=266120735,tx_errors=0,tx_packets=206002 1474485262000000000
>>> Processing stats for 4
> port_stats,host=Switch-VM,port=4,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=12139642,rx_errors=0,rx_packets=69244,tx_bytes=1091206,tx_errors=0,tx_packets=13720 1474485262000000000
>>> Processing stats for 48
> port_stats,host=Switch-VM,port=48,switch=b1668a18-7659-47ac-91ea-bf7357f2dfad rx_bytes=4564410,rx_errors=0,rx_packets=20886,tx_bytes=5523443,tx_errors=0,tx_packets=30435 1474485262000000000
```
