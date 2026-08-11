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
			return nil, fmt.Errorf("empty file returned")
		}
		deviceName := fields[2]
		values := fields[3:]

		if len(values) < 11 {
			return nil, fmt.Errorf("expected at least 11 fields, but got %d", len(values))
		}

		if _, ok := devices[deviceName]; !ok {
			continue
		}

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
