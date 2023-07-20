package cmd

import (
	"context"

	"balance_checker/input"
	"balance_checker/output"
	"balance_checker/proxy"
	"balance_checker/worker"
)

var cfgKey = contextKey("config")

type Config struct {
	ProxyConfig proxy.Config `yaml:"proxy"`

	InputConfig input.Config `yaml:"input"`

	OutputConfig output.Config `yaml:"output"`

	WorkerConfig worker.Config `yaml:"worker"`
}

func configFromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey).(*Config)
}
