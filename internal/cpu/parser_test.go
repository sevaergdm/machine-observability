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
	f, err := os.Open("testdata/stat")
	if err != nil {
		t.Fatalf("unexpected error opening test file: %v", err)
	}
	defer f.Close()

	ts, _ := time.Parse("2006-01-02 15:04:05", tsString)

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
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "all",
					User:    200,
					Nice:    200,
					System:  200,
					Idle:    200,
					Iowait:  200,
					Irq:     200,
					SoftIrq: 200,
					Steal:   200,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "0",
					User:    100,
					Nice:    100,
					System:  100,
					Idle:    100,
					Iowait:  100,
					Irq:     100,
					SoftIrq: 100,
					Steal:   100,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "1",
					User:    100,
					Nice:    100,
					System:  100,
					Idle:    100,
					Iowait:  100,
					Irq:     100,
					SoftIrq: 100,
					Steal:   100,
				},
			},
		},
		{
			name: "Real /proc/stat data",
			input: f,
			want: []Entry{
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "all",
					User:    243226,
					Nice:    14,
					System:  2428805,
					Idle:    56930038,
					Iowait:  7872257,
					Irq:     119958,
					SoftIrq: 70720,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "0",
					User:    2370,
					Nice:    0,
					System:  1139057,
					Idle:    3027774,
					Iowait:  17892,
					Irq:     34891,
					SoftIrq: 16945,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "1",
					User:    21831,
					Nice:    0,
					System:  307636,
					Idle:    3279486,
					Iowait:  561266,
					Irq:     14236,
					SoftIrq: 32572,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "2",
					User:    18206,
					Nice:    1,
					System:  38982,
					Idle:    3625393,
					Iowait:  537499,
					Irq:     7214,
					SoftIrq: 4675,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "3",
					User:    27215,
					Nice:    0,
					System:  359004,
					Idle:    2953690,
					Iowait:  865931,
					Irq:     12893,
					SoftIrq: 4862,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "4",
					User:    10903,
					Nice:    1,
					System:  15418,
					Idle:    3523749,
					Iowait:  680587,
					Irq:     2792,
					SoftIrq: 555,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "5",
					User:    10616,
					Nice:    0,
					System:  14265,
					Idle:    3692863,
					Iowait:  511027,
					Irq:     2770,
					SoftIrq: 469,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "6",
					User:    4942,
					Nice:    5,
					System:  13101,
					Idle:    3968506,
					Iowait:  248808,
					Irq:     1303,
					SoftIrq: 268,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "7",
					User:    11597,
					Nice:    0,
					System:  17818,
					Idle:    3950967,
					Iowait:  244239,
					Irq:     6710,
					SoftIrq: 911,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "8",
					User:    6368,
					Nice:    0,
					System:  6058,
					Idle:    4091412,
					Iowait:  129982,
					Irq:     1557,
					SoftIrq: 223,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "9",
					User:    34901,
					Nice:    0,
					System:  263413,
					Idle:    3220115,
					Iowait:  685787,
					Irq:     10718,
					SoftIrq: 3286,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "10",
					User:    17868,
					Nice:    2,
					System:  40874,
					Idle:    3848860,
					Iowait:  320671,
					Irq:     3391,
					SoftIrq: 829,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "11",
					User:    42732,
					Nice:    0,
					System:  179075,
					Idle:    2316182,
					Iowait:  1661319,
					Irq:     11891,
					SoftIrq: 3304,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "12",
					User:    11875,
					Nice:    0,
					System:  11838,
					Idle:    3607514,
					Iowait:  585615,
					Irq:     2762,
					SoftIrq: 614,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "13",
					User:    9196,
					Nice:    0,
					System:  10358,
					Idle:    3727324,
					Iowait:  479747,
					Irq:     3050,
					SoftIrq: 420,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "14",
					User:    3497,
					Nice:    0,
					System:  2879,
					Idle:    4093405,
					Iowait:  137542,
					Irq:     797,
					SoftIrq: 118,
					Steal:   0,
				},
				{
					BootId:  bootId,
					Ts:      ts,
					Cpu:     "15",
					User:    9102,
					Nice:    0,
					System:  9020,
					Idle:    4002790,
					Iowait:  204336,
					Irq:     2977,
					SoftIrq: 663,
					Steal:   0,
				},
			},

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
