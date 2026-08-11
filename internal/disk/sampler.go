package disk

import (
	"context"
	"fmt"
	"machine-observability/internal/collector"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

type Sampler struct {
	BootId string
}

func (s *Sampler) Name() string { return "disk" }

func (s *Sampler) Sample(ctx context.Context) ([]collector.Event, error) {
	var events []collector.Event
	ts := time.Now().UTC()
	devices, err := fetchDevices()
	if err != nil {
		return nil, err
	}

	diskStats, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, err
	}
	defer func() { _ = diskStats.Close() }()

	diskStatEvents, err := parseDiskStats(diskStats, s.BootId, ts, devices)
	if err != nil {
		return nil, err
	}

	for _, e := range diskStatEvents {
		events = append(events, e)
	}

	mountInfo, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer func() { _ = mountInfo.Close() }()

	parsedMountInfo, err := parseMountInfo(mountInfo)
	if err != nil {
		return nil, err
	}

	for _, mnt := range parsedMountInfo {
		var st unix.Statfs_t
		if err := unix.Statfs(mnt.mount, &st); err != nil {
			return nil, err
		}

		if st.Blocks < st.Bfree {
			return nil, fmt.Errorf("%s: Bfree exceeds total Blocks", mnt.mount)
		}

		entry := FSEntry{
			BootId:     s.BootId,
			Ts:         ts,
			Mount:      mnt.mount,
			FsType:     mnt.fstype,
			Device:     mnt.device,
			SizeBytes:  int64(st.Blocks) * int64(st.Bsize),
			UsedBytes:  (int64(st.Blocks) - int64(st.Bfree)) * int64(st.Bsize),
			FilesTotal: int64(st.Files),
			FilesUsed:  int64(st.Files) - int64(st.Ffree),
		}

		events = append(events, entry)
	}

	return events, nil
}

func fetchDevices() (map[string]bool, error) {
	devices := make(map[string]bool)

	dir, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("encountered an error listing devices: %w", err)
	}

	for _, dirEntry := range dir {
		devices[dirEntry.Name()] = true
	}
	return devices, nil
}
