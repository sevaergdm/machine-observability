package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}

var testNames = map[string]Kind{
	"journal": Streaming,
	"cpu":     Polling,
}

const validConfig = `
data_dir = "/tmp/data"
state_dir = "/tmp/state"

[collectors.journal]
enabled = true

[collectors.cpu]
enabled = true
interval = "10s"
`

func TestLoadValid (t *testing.T) {
	cfg, err := Load(writeConfig(t, validConfig), testNames)
	if err != nil {
		t.Fatalf("unexpected error: %q", err)
	}

	if cfg.DataDir != "/tmp/data" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "/tmp/data")
	}

	if cfg.StateDir != "/tmp/state" {
		t.Errorf("DataDir = %q, want %q", cfg.StateDir, "/tmp/state")
	}

	if !cfg.Collectors["journal"].Enabled {
		t.Error("journal should be enabled")
	}

	if got := cfg.Collectors["cpu"].Interval.Duration; got != 10*time.Second {
		t.Errorf("cput interval = %v, want 10s", got)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "unknown collector name",
			content: `
			data_dir = "/tmp/data"
			state_dir = "/tmp/state"
			[collectors.jornal]
			enabled = true
			`,
			wantErr: "unknown collector",
		},
		{
			name: "misspelled field is rejected",
			content: `
			data_dir = "/tmp/data"
			state_dir = "/tmp/state"
			[collectors.journal]
			enaled = true
			`,
			wantErr: "unknown keys",
		},
  	{
  		name: "interval on streaming collector",
  		content: `
  		data_dir = "/tmp/data"
  		state_dir = "/tmp/state"
  		[collectors.journal]
  		enabled = true
  		interval = "10s"
  		`,
  		wantErr: "streaming",
  	},
		{
			name: "invalid toml syntax",
			content: `
			data_dir = "/tmp/data
			state_dir = "tmp/state"
			[collectors.journal]
			enabled = true
			`,
			wantErr: "strings cannot contain newlines",
		},
		{
			name: "missing data dir",
			content: `
  		state_dir = "/tmp/state"
  		[collectors.journal]
  		enabled = true
			`,
			wantErr: "valid data",
		},
		{
			name: "polling collector missing interval",
  		content: `
  		data_dir = "/tmp/data"
  		state_dir = "/tmp/state"
  		[collectors.cpu]
  		enabled = true
  		`,
			wantErr: "must have a positive interval",
		},
		{
			name: "polling collector missing interval",
  		content: `
  		data_dir = "/tmp/data"
  		state_dir = "/tmp/state"
  		[collectors.cpu]
  		enabled = true
			interval = "-5s"
  		`,
			wantErr: "interval durations must be positive",
		},
		{
			name: "polling collector has invalid interval",
  		content: `
  		data_dir = "/tmp/data"
  		state_dir = "/tmp/state"
  		[collectors.cpu]
  		enabled = true
			interval = "10 seconds"
  		`,
			wantErr: "unknown unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, tt.content), testNames)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
