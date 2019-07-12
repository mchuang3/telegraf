package agent

import (
	"github.com/mchuang3/telegraf/internal/config"
)

// ParseConfig parses the data into a telegraf configuration structure.
func ParseConfig(path string, data []byte) (*config.Config, error) {
	cfg := config.NewConfig()
	err := cfg.ParseConfig(path, data)
	return cfg, err
}
