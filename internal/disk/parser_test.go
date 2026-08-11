package disk

import (
	"io"
	//"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const bootId = "boot-123"
const tsString = "2026-08-06 21:00:00"

var devices = map[string]bool{"nvme0n1": true, "zram0": true}

func TestParseDiskStats(t *testing.T) {
	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	tests := []struct {
		name    string
		input   io.Reader
		want    []IOEntry
		wantErr bool
	}{
		{
			name: "Simple synthetic entry",
			input: strings.NewReader(`
259       0 nvme0n1 109112 13817 8228361 123170 3700810 26007 100662558 5544472 0 528866 5738463 31046 0 66087448 60937 14597 9881
259       1 nvme0n1p1 522 1645 5609 195 637 393 127182 1261 0 94 1457 0 0 0 0 0 0
259       2 nvme0n1p2 108427 12172 8217368 122950 3700168 25614 100535376 5543209 0 618885 5727097 31046 0 66087448 60937 0 0
253       0 zram0 353 0 6072 7 3101 0 37072 113 0 120 120 0 0 0 0 0 0
			`),
			want: []IOEntry{
				{
					BootId:          "boot-123",
					Ts:              ts,
					ReadsCompleted:  int64(109112),
					ReadsMerged:     int64(13817),
					ReadSectors:     int64(8228361),
					ReadMs:          int64(123170),
					WritesCompleted: int64(3700810),
					WritesMerged:    int64(26007),
					WriteSectors:    int64(100662558),
					WriteMs:         int64(5544472),
					IoInProgress:    int64(0),
					IoMs:            int64(528866),
					IoMsWeighted:    int64(5738463),
				},
			},
		},
		{
			name: "error with only 10 fields",
			input: strings.NewReader(`
259       0 nvme0n1 109112 13817 8228361 123170 3700810 26007 100662558 5544472 0 528866
259       1 nvme0n1p1 522 1645 5609 195 637 393 127182 1261 0 94
259       2 nvme0n1p2 108427 12172 8217368 122950 3700168 25614 100535376 5543209 0 618885
253       0 zram0 353 0 6072 7 3101 0 37072 113 0 120			
`),
			wantErr: true,
		},
		{
			name: "error non-numeric value",
			input: strings.NewReader(`
259       0 nvme0n1 aaaaa 13817 8228361 123170 3700810 26007 100662558 5544472 0 528866 5738463 31046 0 66087448 60937 14597 9881
259       1 nvme0n1p1 522 1645 5609 195 637 393 127182 1261 0 94 1457 0 0 0 0 0 0
259       2 nvme0n1p2 108427 12172 8217368 122950 3700168 25614 100535376 5543209 0 618885 5727097 31046 0 66087448 60937 0 0
253       0 zram0 353 0 6072 7 3101 0 37072 113 0 120 120 0 0 0 0 0 0
			`),
			wantErr: true,
		},
		{
			name:    "no lines, expect error",
			input:   strings.NewReader(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDiskStats(tt.input, bootId, ts, devices)

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

/*
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
*/
