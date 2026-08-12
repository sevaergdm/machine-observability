package network

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func parseNetDev(r io.Reader, bootId string, ts time.Time) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		iface, values, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		iface = strings.TrimSpace(iface)

		splitValues := strings.Fields(values)
		if len(splitValues) < 16 {
			return nil, fmt.Errorf("incomplete data for iface '%s', expected at least 16 values, but got %d", iface, len(splitValues))
		}

		var parseErr error
		p := func(i int) int64 {
			v, err := strconv.ParseInt(splitValues[i], 10, 64)
			if err != nil && parseErr == nil {
				parseErr = fmt.Errorf("value %d (%q): %w", i, splitValues[i], err)
			}
			return v
		}

		entry := Entry{
			BootId:    bootId,
			Ts:        ts,
			Interface: iface,
			RxBytes:   p(0),
			RxPackets: p(1),
			RxErrs:    p(2),
			RxDrop:    p(3),
			TxBytes:   p(8),
			TxPackets: p(9),
			TxErrs:    p(10),
			TxDrop:    p(11),
		}

		if parseErr != nil {
			return nil, parseErr
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("encountered an error reading /proc/net/dev: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no lines found in /proc/net/dev")
	}

	return entries, nil
}
