package agent

import (
	"fmt"
	"log"
	"math"
	"sync/atomic"
	"time"

	"github.com/mchuang3/telegraf"
	"github.com/mchuang3/telegraf/internal/models"
)

func NewAccumulator(
	inputConfig *models.InputConfig,
	metrics chan telegraf.Metric,
) *accumulator {
	acc := accumulator{}
	acc.metrics = metrics
	acc.inputConfig = inputConfig
	acc.precision = time.Nanosecond
	return &acc
}

type accumulator struct {
	metrics chan telegraf.Metric

	defaultTags map[string]string

	debug bool
	// print every point added to the accumulator
	trace bool

	inputConfig *models.InputConfig

	precision time.Duration

	errCount uint64
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if m := ac.makeMetric(measurement, fields, tags, telegraf.Untyped, t...); m != nil {
		ac.metrics <- m
	}
}

func (ac *accumulator) AddGauge(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if m := ac.makeMetric(measurement, fields, tags, telegraf.Gauge, t...); m != nil {
		ac.metrics <- m
	}
}

func (ac *accumulator) AddCounter(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if m := ac.makeMetric(measurement, fields, tags, telegraf.Counter, t...); m != nil {
		ac.metrics <- m
	}
}

// makeMetric either returns a metric, or returns nil if the metric doesn't
// need to be created (because of filtering, an error, etc.)
func (ac *accumulator) makeMetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	mType telegraf.ValueType,
	t ...time.Time,
) telegraf.Metric {
	if len(fields) == 0 || len(measurement) == 0 {
		return nil
	}
	if tags == nil {
		tags = make(map[string]string)
	}

	// Override measurement name if set
	if len(ac.inputConfig.NameOverride) != 0 {
		measurement = ac.inputConfig.NameOverride
	}
	// Apply measurement prefix and suffix if set
	if len(ac.inputConfig.MeasurementPrefix) != 0 {
		measurement = ac.inputConfig.MeasurementPrefix + measurement
	}
	if len(ac.inputConfig.MeasurementSuffix) != 0 {
		measurement = measurement + ac.inputConfig.MeasurementSuffix
	}

	// Apply plugin-wide tags if set
	for k, v := range ac.inputConfig.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	// Apply daemon-wide tags if set
	for k, v := range ac.defaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	// Apply the metric filter(s)
	if ok := ac.inputConfig.Filter.Apply(measurement, fields, tags); !ok {
		return nil
	}

	for k, v := range fields {
		// Validate uint64 and float64 fields
		switch val := v.(type) {
		case uint64:
			// InfluxDB does not support writing uint64
			if val < uint64(9223372036854775808) {
				fields[k] = int64(val)
			} else {
				fields[k] = int64(9223372036854775807)
			}
			continue
		case float64:
			// NaNs are invalid values in influxdb, skip measurement
			if math.IsNaN(val) || math.IsInf(val, 0) {
				if ac.debug {
					log.Printf("Measurement [%s] field [%s] has a NaN or Inf "+
						"field, skipping",
						measurement, k)
				}
				delete(fields, k)
				continue
			}
		}

		fields[k] = v
	}

	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	timestamp = timestamp.Round(ac.precision)

	var m telegraf.Metric
	var err error
	switch mType {
	case telegraf.Counter:
		m, err = telegraf.NewCounterMetric(measurement, tags, fields, timestamp)
	case telegraf.Gauge:
		m, err = telegraf.NewGaugeMetric(measurement, tags, fields, timestamp)
	default:
		m, err = telegraf.NewMetric(measurement, tags, fields, timestamp)
	}
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return nil
	}

	if ac.trace {
		fmt.Println("> " + m.String())
	}

	return m
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	atomic.AddUint64(&ac.errCount, 1)
	//TODO suppress/throttle consecutive duplicate errors?
	log.Printf("ERROR in input [%s]: %s", ac.inputConfig.Name, err)
}

func (ac *accumulator) Debug() bool {
	return ac.debug
}

func (ac *accumulator) SetDebug(debug bool) {
	ac.debug = debug
}

func (ac *accumulator) Trace() bool {
	return ac.trace
}

func (ac *accumulator) SetTrace(trace bool) {
	ac.trace = trace
}

// SetPrecision takes two time.Duration objects. If the first is non-zero,
// it sets that as the precision. Otherwise, it takes the second argument
// as the order of time that the metrics should be rounded to, with the
// maximum being 1s.
func (ac *accumulator) SetPrecision(precision, interval time.Duration) {
	if precision > 0 {
		ac.precision = precision
		return
	}
	switch {
	case interval >= time.Second:
		ac.precision = time.Second
	case interval >= time.Millisecond:
		ac.precision = time.Millisecond
	case interval >= time.Microsecond:
		ac.precision = time.Microsecond
	default:
		ac.precision = time.Nanosecond
	}
}

func (ac *accumulator) DisablePrecision() {
	ac.precision = time.Nanosecond
}

func (ac *accumulator) setDefaultTags(tags map[string]string) {
	ac.defaultTags = tags
}

func (ac *accumulator) addDefaultTag(key, value string) {
	if ac.defaultTags == nil {
		ac.defaultTags = make(map[string]string)
	}
	ac.defaultTags[key] = value
}
