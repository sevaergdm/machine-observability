package cpu

import (
	"bufio"
	"errors"
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
		if err := scanner.Err(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("encountered an error reading /proc/stat: %w", err)
		}

		line := scanner.Text()

		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		splitLine := strings.Fields(line)

		var cpu string
		if splitLine[0] == "cpu" {
			cpu = "all"
		} else {
			cpu = strings.TrimPrefix(splitLine[0], "cpu")
		}

		numEntryFields := make([]int64, 7)
		for i := 1; i < 8; i++ {
			tmp, err := strconv.ParseInt(splitLine[i], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("encountered an error converting value to int: %w", err)
			}
			numEntryFields = append(numEntryFields, tmp)
		}

		entry := Entry{
			BootId:  bootId,
			Ts:      ts,
			Cpu:     cpu,
			User:    numEntryFields[0],
			Nice:    numEntryFields[1],
			System:  numEntryFields[2],
			Idle:    numEntryFields[3],
			Iowait:  numEntryFields[4],
			Irq:     numEntryFields[5],
			SoftIrq: numEntryFields[6],
		}

		entries = append(entries, entry)
	}
	return entries, nil
}
