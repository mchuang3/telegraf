package influx

import (
	"github.com/mchuang3/telegraf"
)

type InfluxSerializer struct {
}

func (s *InfluxSerializer) Serialize(m telegraf.Metric) ([]byte, error) {
	return m.Serialize(), nil
}
