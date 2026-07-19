package config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Kind int

const (
	Streaming Kind = iota
	Polling
)

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	d.Duration = v
	return err
}

type CollectorConfig struct {
	Enabled  bool     `toml:"enabled"`
	Interval Duration `toml:"interval"`
}

type Config struct {
	DataDir    string                     `toml:"data_dir"`
	StateDir   string                     `toml:"state_dir"`
	Collectors map[string]CollectorConfig `toml:"collectors"`
}

func Validate(cfg Config, validNames map[string]Kind) error {
	if cfg.DataDir == "" {
		return fmt.Errorf("a valid data directory path must be provided")
	}

	if cfg.StateDir == "" {
		return fmt.Errorf("a valid state directory path must be provided")
	}

	for name, value := range cfg.Collectors {
		kind, ok := validNames[name]
		if !ok {
			return fmt.Errorf("unknown collector %q in config", name)
		}

		switch kind {
		case Polling:
			if value.Interval.Duration == 0 {
				return fmt.Errorf("polling sources must have a positive interval set")
			}
			if value.Interval.Duration < 0 {
				return fmt.Errorf("interval durations must be positive, got: %v", value.Interval)
			}
		case Streaming:
			if value.Interval.Duration != 0 {
				return fmt.Errorf("streaming sources cannot have an interval set")
			}
		}
	}
	return nil
}

func Load(path string, validNames map[string]Kind) (Config, error) {
	var cfg Config
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, err
	}

	missingKeys := meta.Undecoded()
	if len(missingKeys) > 0 {
		return Config{}, fmt.Errorf("unknown keys in config file: %+v", missingKeys)
	}

	err = Validate(cfg, validNames)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
