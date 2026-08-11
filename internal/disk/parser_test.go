package disk

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
					DeviceName:      "nvme0n1",
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
				{
					BootId:          "boot-123",
					Ts:              ts,
					DeviceName:      "zram0",
					ReadsCompleted:  int64(353),
					ReadsMerged:     int64(0),
					ReadSectors:     int64(6072),
					ReadMs:          int64(7),
					WritesCompleted: int64(3101),
					WritesMerged:    int64(0),
					WriteSectors:    int64(37072),
					WriteMs:         int64(113),
					IoInProgress:    int64(0),
					IoMs:            int64(120),
					IoMsWeighted:    int64(120),
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

func TestParseDiskStatsRealFile(t *testing.T) {
	f, err := os.Open("testdata/diskstats")
	if err != nil {
		t.Fatalf("unexpected error opening test file: %v", err)
	}
	defer func() { _ = f.Close() }()

	ts, err := time.Parse("2006-01-02 15:04:05", tsString)
	if err != nil {
		t.Fatalf("unexpected error parsing timestamp '%s': %v", tsString, err)
	}

	got, err := parseDiskStats(f, bootId, ts, devices)
	if err != nil {
		t.Fatalf("unexpected error parsing %s: %v", f.Name(), err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, but got %d", len(got))
	}

	countDevices := make(map[string]int)
	for _, v := range got {
		countDevices[v.DeviceName]++

		// every busy ms adds 1 to io_ms and at least 1 to weighted. queue depth >= 1 whenever busy
		if v.IoMsWeighted < v.IoMs {
			t.Errorf("%s: IoMsWeighted (%d) must match or exceed IoMs (%d)", v.DeviceName, v.IoMsWeighted, v.IoMs)
		}

		// column shift tripwire for the only gauge column. It's value is most often 0 but should never really exceed a few hundred
		if v.IoInProgress > 10000 {
			t.Errorf("%s: expect IoInProgress to be a low value, but got %d", v.DeviceName, v.IoInProgress)
		}

		fields := []struct {
			name  string
			value int64
		}{
			{"reads_completed", v.ReadsCompleted},
			{"reads_merged", v.ReadsMerged},
			{"read_sectors", v.ReadSectors},
			{"read_ms", v.ReadMs},
			{"writes_completed", v.WritesCompleted},
			{"writes_merged", v.WritesMerged},
			{"write_sectors", v.WriteSectors},
			{"write_ms", v.WriteMs},
			{"io_in_progress", v.IoInProgress},
			{"io_ms", v.IoMs},
			{"io_ms_weighted", v.IoMsWeighted},
		}

		for _, field := range fields {
			if field.value < 0 {
				t.Errorf("%s: %s = %d, want >= 0", v.DeviceName, field.name, field.value)
			}

		}
	}

	for _, want := range []string{"nvme0n1", "zram0"} {
		if countDevices[want] != 1 {
			t.Errorf("device %s: %d entries, want exactly 1", want, countDevices[want])
		}
	}
}

func TestParseMountInfo(t *testing.T) {
	tests := []struct {
		name    string
		input   io.Reader
		want    []mountEntry
		wantErr bool
	}{
		{
			name: "Simple entry",
			input: strings.NewReader(`
31 2 0:28 /@ / rw,relatime shared:1 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=256,subvol=/@
52 31 0:28 /@home /home rw,relatime shared:111 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=257,subvol=/@home
54 31 0:28 /@pkg /var/cache/pacman/pkg rw,relatime shared:115 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=259,subvol=/@pkg
57 31 0:28 /@log /var/log rw,relatime shared:119 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=258,subvol=/@log
64 31 259:1 / /boot rw,relatime shared:123 - vfat /dev/nvme0n1p1 rw,fmask=0077,dmask=0077,codepage=437,iocharset=ascii,shortname=mixed,utf8,errors=remount-ro
			`),
			want: []mountEntry{
				{
					mount:  "/",
					fstype: "btrfs",
					device: "/dev/nvme0n1p2",
				},
				{
					mount:  "/home",
					fstype: "btrfs",
					device: "/dev/nvme0n1p2",
				},
				{
					mount:  "/var/cache/pacman/pkg",
					fstype: "btrfs",
					device: "/dev/nvme0n1p2",
				},
				{
					mount:  "/var/log",
					fstype: "btrfs",
					device: "/dev/nvme0n1p2",
				},
				{
					mount:  "/boot",
					fstype: "vfat",
					device: "/dev/nvme0n1p1",
				},
			},
		},
		{
			name: "incomplete entry",
			input: strings.NewReader(`
31 2 0:28 /@ 
52 31 0:28 /@home 
54 31 0:28 /@pkg /var/cache/pacman/pkg rw,relatime shared:115 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=259,subvol=/@pkg
57 31 0:28 /@log /var/log rw,relatime shared:119 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=258,subvol=/@log
64 31 259:1 / /boot rw,relatime shared:123 - vfat /dev/nvme0n1p1 rw,fmask=0077,dmask=0077,codepage=437,iocharset=ascii,shortname=mixed,utf8,errors=remount-ro
			`),
			wantErr: true,
		},
		{
			name: "missing separator",
			input: strings.NewReader(`
31 2 0:28 /@ / rw,relatime shared:1  btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=256,subvol=/@
52 31 0:28 /@home /home rw,relatime shared:111 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=257,subvol=/@home
54 31 0:28 /@pkg /var/cache/pacman/pkg rw,relatime shared:115 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=259,subvol=/@pkg
57 31 0:28 /@log /var/log rw,relatime shared:119 - btrfs /dev/nvme0n1p2 rw,compress=zstd:3,ssd,discard=async,space_cache=v2,subvolid=258,subvol=/@log
64 31 259:1 / /boot rw,relatime shared:123 - vfat /dev/nvme0n1p1 rw,fmask=0077,dmask=0077,codepage=437,iocharset=ascii,shortname=mixed,utf8,errors=remount-ro
			`),
			wantErr: true,
		},
		{
			name:    "empty file",
			input:   strings.NewReader(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMountInfo(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(mountEntry{})); diff != "" {
				t.Errorf("Parse mismatch: %v", diff)
			}
		})
	}
}

func TestParseMountInfoRealFile(t *testing.T) {
	f, err := os.Open("testdata/mountinfo")
	if err != nil {
		t.Fatalf("unexpected error opening test file: %v", err)
	}
	defer func() { _ = f.Close() }()

	got, err := parseMountInfo(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 5 {
		t.Errorf("expected 5 entries, but got %d", len(got))
	}

	countMounts := make(map[string]int)
	for _, v := range got {
		countMounts[v.mount]++

		if !strings.HasPrefix(v.device, "/dev/") {
			t.Errorf("device %s: missing /dev/ prefix", v.device)
		}

		if !strings.HasPrefix(v.mount, "/") {
			t.Errorf("mount %s: missing '/' prefix", v.mount)
		}

		if v.fstype == "" || v.fstype == "-" || strings.Contains(v.fstype, ":") {
			t.Errorf("fstype %s is missing or malformed", v.fstype)
		}
	}

	if countMounts["/"] == 0 {
		t.Errorf("missing root mount in file")
	}

	for mount, count := range countMounts {
		if count > 1 {
			t.Errorf("%s: has %d entries, should only have 1", mount, count)
		}
	}
}
