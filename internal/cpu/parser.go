package cpu

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func parseStat(r io.Reader, bootId string, ts time.Time) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		splitLine := strings.Fields(line)
		if len(splitLine) < 9 {
			return nil, fmt.Errorf("expected at least 9 fields, but got %d", len(splitLine))
		}

		var cpu string
		if splitLine[0] == "cpu" {
			cpu = "all"
		} else {
			cpu = strings.TrimPrefix(splitLine[0], "cpu")
		}

		// using a closure to convert all numeric fields to int64 and capturing any errors
		// it doesn't matter if parseErr is overwritten by any particular value because we will fail and
		// return on any error
		var parseErr error
		p := func(i int) int64 {
			v, err := strconv.ParseInt(splitLine[i], 10, 64)
			if err != nil && parseErr == nil {
				parseErr = fmt.Errorf("value %d (%q): %w", i, splitLine[i], err)
			}
			return v
		}

		entry := Entry{
			BootId:  bootId,
			Ts:      ts,
			Cpu:     cpu,
			User:    p(1),
			Nice:    p(2),
			System:  p(3),
			Idle:    p(4),
			Iowait:  p(5),
			Irq:     p(6),
			SoftIrq: p(7),
			Steal:   p(8),
		}
		if parseErr != nil {
			return nil, parseErr
		}

		if scanner.Err() != nil {
			return nil, fmt.Errorf("encountered an error reading /proc/stat: %w", scanner.Err())
		}

		entries = append(entries, entry)
		if entries == nil {
			return nil, fmt.Errorf("returned no rows after parsing")
		}
	}
	return entries, nil
}
