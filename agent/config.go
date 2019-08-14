package agent

import (
	"github.com/mchuang3/telegraf/internal/config"

	_ "github.com/mchuang3/telegraf/plugins/aggregators/all"
	_ "github.com/mchuang3/telegraf/plugins/inputs/all"
	_ "github.com/mchuang3/telegraf/plugins/outputs/all"
	_ "github.com/mchuang3/telegraf/plugins/processors/all"
)

// ParseConfig parses the data into a telegraf configuration structure.
func ParseConfig(path string, data []byte) (*config.Config, error) {
	cfg := config.NewConfig()
	err := cfg.ParseConfig(path, data)
	return cfg, err
}
