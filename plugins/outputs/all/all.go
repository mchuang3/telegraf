package all

import (
	_ "github.com/mchuang3/telegraf/plugins/outputs/amon"
	_ "github.com/mchuang3/telegraf/plugins/outputs/amqp"
	_ "github.com/mchuang3/telegraf/plugins/outputs/cloudwatch"
	_ "github.com/mchuang3/telegraf/plugins/outputs/datadog"
	_ "github.com/mchuang3/telegraf/plugins/outputs/file"
	_ "github.com/mchuang3/telegraf/plugins/outputs/graphite"
	_ "github.com/mchuang3/telegraf/plugins/outputs/graylog"
	_ "github.com/mchuang3/telegraf/plugins/outputs/influxdb"
	_ "github.com/mchuang3/telegraf/plugins/outputs/instrumental"
	_ "github.com/mchuang3/telegraf/plugins/outputs/kafka"
	_ "github.com/mchuang3/telegraf/plugins/outputs/kinesis"
	_ "github.com/mchuang3/telegraf/plugins/outputs/librato"
	_ "github.com/mchuang3/telegraf/plugins/outputs/mqtt"
	_ "github.com/mchuang3/telegraf/plugins/outputs/nats"
	_ "github.com/mchuang3/telegraf/plugins/outputs/nsq"
	_ "github.com/mchuang3/telegraf/plugins/outputs/opentsdb"
	_ "github.com/mchuang3/telegraf/plugins/outputs/prometheus_client"
	_ "github.com/mchuang3/telegraf/plugins/outputs/riemann"
)
