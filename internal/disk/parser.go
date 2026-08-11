package disk

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func parseDiskStats(r io.Reader, bootId string, ts time.Time, devices map[string]bool) ([]IOEntry, error) {
	var entries []IOEntry

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if len(fields) < 3 {
			return nil, fmt.Errorf("expected at least 3 fields, but got %d", len(fields))
		}

		deviceName := fields[2]
		if !devices[deviceName] {
			continue
		}

		if len(fields) < 14 {
			return nil, fmt.Errorf("device %s: expected at least 11 stat values, but got %d", deviceName, len(fields)-3)
		}
		values := fields[3:]

		// capture the first parse error; later calls still run but don't mask
		var parseErr error
		p := func(i int) int64 {
			v, err := strconv.ParseInt(values[i], 10, 64)
			if err != nil && parseErr == nil {
				parseErr = fmt.Errorf("value %d (%q): %w", i, values[i], err)
			}
			return v
		}

		entry := IOEntry{
			BootId:          bootId,
			Ts:              ts,
			DeviceName:      deviceName,
			ReadsCompleted:  p(0),
			ReadsMerged:     p(1),
			ReadSectors:     p(2),
			ReadMs:          p(3),
			WritesCompleted: p(4),
			WritesMerged:    p(5),
			WriteSectors:    p(6),
			WriteMs:         p(7),
			IoInProgress:    p(8),
			IoMs:            p(9),
			IoMsWeighted:    p(10),
		}

		if parseErr != nil {
			return nil, parseErr
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("encountered an error reading /proc/diskstats: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no diskstats lines found")
	}
	return entries, nil
}

type mountEntry struct {
	mount  string
	fstype string
	device string
}

func parseMountInfo(r io.Reader) ([]mountEntry, error) {
	var entries []mountEntry
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if len(fields) < 5 {
			return nil, fmt.Errorf("expected at least 5 fields, but got %d", len(fields))
		}

		mount := fields[4]
		separatorIndex := -1
		for i, field := range fields {
			if field == "-" {
				separatorIndex = i
				break
			}
		}

		if separatorIndex == -1 {
			return nil, fmt.Errorf("malformed file, no '-' separator found")
		}

		if separatorIndex+2 >= len(fields) {
			return nil, fmt.Errorf("malformed line: missing fstype/source after separator")
		}

		fstype := fields[separatorIndex+1]
		device := fields[separatorIndex+2]

		if !strings.HasPrefix(device, "/dev/") {
			continue
		}

		entries = append(entries, mountEntry{mount: mount, fstype: fstype, device: device})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("encountered an error reading /proc/self/mountinfo: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no lines found in mountinfo")
	}
	return entries, nil
}
