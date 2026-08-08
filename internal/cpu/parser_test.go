package cpu

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

func TestParseStat(t *testing.T) {
	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	tests := []struct {
		name    string
		input   io.Reader
		want    []Entry
		wantErr bool
	}{
		{
			name: "Simple synthetic entry",
			input: strings.NewReader(`
cpu 200 200 200 200 200 200 200 200
cpu0 100 100 100 100 100 100 100 100
cpu1 100 100 100 100 100 100 100 100
intr 12345
			`),
			want: []Entry{
				{BootId: bootId, Ts: ts, Cpu: "all", User: 200, Nice: 200, System: 200, Idle: 200, Iowait: 200, Irq: 200, SoftIrq: 200, Steal: 200},
				{BootId: bootId, Ts: ts, Cpu: "0", User: 100, Nice: 100, System: 100, Idle: 100, Iowait: 100, Irq: 100, SoftIrq: 100, Steal: 100},
				{BootId: bootId, Ts: ts, Cpu: "1", User: 100, Nice: 100, System: 100, Idle: 100, Iowait: 100, Irq: 100, SoftIrq: 100, Steal: 100},
			},
		},
		{
			name: "error with only 7 fields",
			input: strings.NewReader(`
cpu 200 200 200 200 200 200 200
cpu0 100 100 100 100 100 100 100
cpu1 100 100 100 100 100 100 100
intr 12345
			`),
			wantErr: true,
		},
		{
			name: "error non-numeric jiffy",
			input: strings.NewReader(`
cpu aaa 200 200 200 200 200 200 200
cpu0 100 100 100 100 100 100 100 100
cpu1 100 100 100 100 100 100 100 100
intr 12345
			`),
			wantErr: true,
		},
		{
			name: "Synthetic entry with more fields, should parse and ignore extras",
			input: strings.NewReader(`
cpu 200 200 200 200 200 200 200 200 200 200 200
cpu0 100 100 100 100 100 100 100 100 100 100 100
cpu1 100 100 100 100 100 100 100 100 100 100 100
intr 12345
			`),
			want: []Entry{
				{BootId: bootId, Ts: ts, Cpu: "all", User: 200, Nice: 200, System: 200, Idle: 200, Iowait: 200, Irq: 200, SoftIrq: 200, Steal: 200},
				{BootId: bootId, Ts: ts, Cpu: "0", User: 100, Nice: 100, System: 100, Idle: 100, Iowait: 100, Irq: 100, SoftIrq: 100, Steal: 100},
				{BootId: bootId, Ts: ts, Cpu: "1", User: 100, Nice: 100, System: 100, Idle: 100, Iowait: 100, Irq: 100, SoftIrq: 100, Steal: 100},
			},
		},
		{
			name:    "no cpu lines, expect error",
			input:   strings.NewReader("intr 12345"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, _ := time.Parse("2006-01-02 15:04:05", tsString)
			got, err := parseStat(tt.input, bootId, ts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got none")
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

func TestParseStatRealFile(t *testing.T) {
	f, err := os.Open("testdata/stat")
	if err != nil {
		t.Fatalf("unexpected error opening test file: %v", err)
	}
	defer func() { _ = f.Close() }()

	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	got, err := parseStat(f, bootId, ts)
	if err != nil {
		t.Fatalf("unexpected error parsing %s: %v", f.Name(), err)
	}

	if len(got) != 17 {
		t.Fatalf("expected 17 entries, but got %d", len(got))
	}

	var sum, all Entry
	for _, entry := range got {
		if entry.Cpu == "all" {
			all = entry
			continue
		}
		sum.User += entry.User
		sum.Nice += entry.Nice
		sum.System += entry.System
		sum.Idle += entry.Idle
		sum.Iowait += entry.Iowait
		sum.Irq += entry.Irq
		sum.SoftIrq += entry.SoftIrq
		sum.Steal += entry.Steal
	}

	if all.Cpu != "all" {
		t.Fatalf("no aggregate 'all' row in parsed output")
	}

	// Because the total sum on the cpu line is calculated prior to truncation to "jiffies" by the kernel the value can differ by up to 1 jiffy
	// To accommodate this we account for a max diff being the total number of CPUs on the machine and expect the gap to never be negative due to the floor division
	numCpus := len(got) - 1
	checks := []struct {
		name string
		all  int64
		sum  int64
	}{
		{"user", all.User, sum.User},
		{"nice", all.Nice, sum.Nice},
		{"system", all.System, sum.System},
		{"idle", all.Idle, sum.Idle},
		{"iowait", all.Iowait, sum.Iowait},
		{"irq", all.Irq, sum.Irq},
		{"softirq", all.SoftIrq, sum.SoftIrq},
		{"steal", all.Steal, sum.Steal},
	}

	for _, check := range checks {
		gap := check.all - check.sum
		if gap < 0 || gap > int64(numCpus) {
			t.Errorf("%s: aggregate-sum = %d, want within [0, %d]", check.name, gap, numCpus)
		}
	}

}
