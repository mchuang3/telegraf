package models

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
)

type RunningAggregator struct {
	Aggregator telegraf.Aggregator
	Config     *AggregatorConfig
}

// FilterConfig containing a name and filter
type AggregatorConfig struct {
	Name   string
	Filter Filter
}

func (ra *RunningAggregator) Name() string {
	return "aggregators." + ra.Config.Name
}

func (ra *RunningAggregator) MakeMetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	mType telegraf.ValueType,
	t time.Time,
) telegraf.Metric {
	var m telegraf.Metric
	var err error
	switch mType {
	case telegraf.Counter:
		m, err = telegraf.NewCounterMetric(measurement, tags, fields, t)
	case telegraf.Gauge:
		m, err = telegraf.NewGaugeMetric(measurement, tags, fields, t)
	default:
		m, err = telegraf.NewMetric(measurement, tags, fields, t)
	}
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return nil
	}

	m.SetAggregate(true)

	return m
}

func (ra *RunningAggregator) Apply(in telegraf.Metric) {
	if ra.Config.Filter.IsActive() {
		// check if the aggregator should apply this metric
		name := in.Name()
		fields := in.Fields()
		tags := in.Tags()
		t := in.Time()
		if ok := ra.Config.Filter.Apply(name, fields, tags); !ok {
			return
		}

		in, _ = telegraf.NewMetric(name, tags, fields, t)
	}

	ra.Aggregator.Apply(in)
}
