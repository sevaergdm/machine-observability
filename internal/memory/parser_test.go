package memory

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const bootId = "boot-123"
const tsString = "2026-08-06 21:00:00"

func TestParseMemInfo(t *testing.T) {
	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	tests := []struct {
		name    string
		input   io.Reader
		want    Entry
		wantErr bool
	}{
		{
			name: "simple synthetic entry",
			input: strings.NewReader(`
MemTotal:       32094844 kB
MemFree:        20195568 kB
MemAvailable:   27233332 kB
Buffers:            2728 kB
Cached:          7083384 kB
SwapCached:            0 kB
SwapTotal:      38203168 kB
SwapFree:       38203168 kB
Inactive:        5830180 kB
Active(anon):    2006124 kB
Inactive(anon):   653764 kB
Active(file):    1814192 kB
Dirty:  				    3912 kB
Slab:             775948 kB
			`),
			want: Entry{
				BootId:    bootId,
				Ts:        ts,
				Total:     int64(32094844) * 1024,
				Free:      int64(20195568) * 1024,
				Available: int64(27233332) * 1024,
				Buffers:   int64(2728) * 1024,
				Cached:    int64(7083384) * 1024,
				SwapTotal: int64(38203168) * 1024,
				SwapFree:  int64(38203168) * 1024,
				Dirty:     int64(3912) * 1024,
				Slab:      int64(775948) * 1024,
			},
		},
		{
			name: "missing target fields",
			input: strings.NewReader(`
MemTotal:       32094844 kB
MemAvailable:   27233332 kB
Buffers:            2728 kB
Cached:          7083384 kB
SwapCached:            0 kB
SwapTotal:      38203168 kB
SwapFree:       38203168 kB
Inactive:        5830180 kB
Active(anon):    2006124 kB
Inactive(anon):   653764 kB
Active(file):    1814192 kB
Slab:             775948 kB
			`),
			wantErr: true,
		},
		{
			name: "non-numeric value in file",
			input: strings.NewReader(`
MemTotal:      32094844a kB
MemFree:        20195568 kB
MemAvailable:   27233332 kB
Buffers:            2728 kB
Cached:          7083384 kB
SwapCached:            0 kB
SwapTotal:      38203168 kB
SwapFree:       38203168 kB
Inactive:        5830180 kB
Active(anon):    2006124 kB
Inactive(anon):   653764 kB
Active(file):    1814192 kB
Dirty:  				    3912 kB
Slab:             775948 kB
			`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemInfo(tt.input, bootId, ts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("Parse mismatch: %v", diff)
			}
		})
	}
}

func TestParseMemInfoRealFile(t *testing.T) {
	f, err := os.Open("testdata/meminfo")
	if err != nil {
		t.Fatalf("unexpected error opening test file: %v", err)
	}
	defer func() { _ = f.Close() }()

	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	got, err := parseMemInfo(f, bootId, ts)
	if err != nil {
		t.Fatalf("unexpected error parsing %s: %v", f.Name(), err)
	}

	if got.Free > got.Total {
		t.Errorf("free memory exceeds total memory")
	}

	if got.Available > got.Total {
		t.Errorf("available memory exceeds total memory")
	}

	if got.SwapFree > got.SwapTotal {
		t.Errorf("free swap memory exceeds total swap memory")
	}

	if got.Buffers+got.Cached > got.Total {
		t.Errorf("buffers + cached exceeds total memory")
	}
}
