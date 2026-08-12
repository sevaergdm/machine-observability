package network

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

func TestParseNetDev(t *testing.T) {
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
			name: "simple synthetic entry",
			input: strings.NewReader(`
Inter-|   Receive                                                |  Transmit
face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
wlan0: 887728761  814034    0    0    0     0          0         0 76287107  213526    0    0    0     0       0          0
			`),
			want: []Entry{
				{
					BootId:    bootId,
					Ts:        ts,
					Interface: "wlan0",
					RxBytes:   int64(887728761),
					RxPackets: int64(814034),
					RxErrs:    int64(0),
					RxDrop:    int64(0),
					TxBytes:   int64(76287107),
					TxPackets: int64(213526),
					TxErrs:    int64(0),
					TxDrop:    int64(0),
				},
			},
		},
		{
			name: "missing values",
			input: strings.NewReader(`
Inter-|   Receive                                                |  Transmit
face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
wlan0: 887728761  814034    0    0    0     0          0         0 76287107  213526    0    0    0 
			`),
			wantErr: true,
		},
		{
			name: "non-numeric value",
			input: strings.NewReader(`
Inter-|   Receive                                                |  Transmit
face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
wlan0: 887728761abcd  814034    0    0    0     0          0         0 76287107  213526    0    0    0 0 0 0 
			`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNetDev(tt.input, bootId, ts)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error, but got: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Parse mismatch: %v", diff)
			}
		})
	}
}

func TestParseNetDevRealFile(t *testing.T) {
	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	f, err := os.Open("testdata/dev")
	if err != nil {
		t.Fatalf("unexpected error parsing %s: %v", f.Name(), err)
	}
	defer func() { _ = f.Close() }()

	got, err := parseNetDev(f, bootId, ts)
	if err != nil {
		t.Errorf("unexpected error parsing file: %v", err)
	}

	if len(got) != 4 {
		t.Fatalf("expected 2 entries, but got %d", len(got))
	}

	countInterfaces := make(map[string]int)
	for _, v := range got {
		countInterfaces[v.Interface]++

		if v.RxBytes < v.RxPackets {
			t.Errorf("RxBytes (%d) should always match or exceed RxPackets (%d)", v.RxBytes, v.RxPackets)
		}

		if v.TxBytes < v.TxPackets {
			t.Errorf("TxBytes (%d) should always match or exceed TxPackets (%d)", v.TxBytes, v.TxPackets)
		}

		fields := []struct {
			name  string
			value int64
		}{
			{"rx_bytes", v.RxBytes},
			{"rx_packets", v.RxPackets},
			{"rx_errs", v.RxErrs},
			{"rx_drop", v.RxDrop},
			{"tx_bytes", v.TxBytes},
			{"tx_packets", v.TxPackets},
			{"tx_errs", v.TxErrs},
			{"tx_drop", v.TxDrop},
		}

		for _, field := range fields {
			if field.value < 0 {
				t.Errorf("%s: %s = %d, want >= 0", v.Interface, field.name, field.value)
			}
		}
	}

	for _, want := range []string{"lo", "wlan0", "enp4s0", "docker0"} {
		if countInterfaces[want] != 1 {
			t.Errorf("device %s: %d entries, want exactly 1", want, countInterfaces[want])
		}
	}
}
