package memory

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func parseMemInfo(r io.Reader, bootId string, ts time.Time) (Entry, error) {
	scanner := bufio.NewScanner(r)
	entry := Entry{BootId: bootId, Ts: ts}
	targets := map[string]*int64{
		"MemTotal":     &entry.Total,
		"MemFree":      &entry.Free,
		"MemAvailable": &entry.Available,
		"Buffers":      &entry.Buffers,
		"Cached":       &entry.Cached,
		"SwapTotal":    &entry.SwapTotal,
		"SwapFree":     &entry.SwapFree,
		"Dirty":        &entry.Dirty,
		"Slab":         &entry.Slab,
	}

	for scanner.Scan() {
		line := scanner.Text()

		splitLine := strings.Fields(line)
		if len(splitLine) == 0 {
			continue
		}

		field := strings.TrimSuffix(splitLine[0], ":")
		ptr, ok := targets[field]
		if !ok {
			continue
		}

		valueKb, err := strconv.ParseInt(splitLine[1], 10, 64)
		if err != nil {
			return Entry{}, fmt.Errorf("unable to parse value for %s: %w", field, err)
		}
		*ptr = valueKb * 1024
		delete(targets, field)
	}

	if err := scanner.Err(); err != nil {
		return Entry{}, fmt.Errorf("encountered an error reading /proc/meminfo: %w", err)
	}

	if len(targets) > 0 {
		remaining := make([]string, 0, len(targets))
		for k := range targets {
			remaining = append(remaining, k)
		}
		return Entry{}, fmt.Errorf("not all targets were parsed: %+v", remaining)
	}

	return entry, nil
}
