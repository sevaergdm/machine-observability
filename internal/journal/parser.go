package journal

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func Parse(raw map[string]any) (JournalEntry, error) {
	realtimeTimestamp, err := parseTimestamp(raw["__REALTIME_TIMESTAMP"])
	if err != nil {
		return JournalEntry{}, err
	}

	monotonicTimestamp, err := parseTimestamp(raw["__MONOTONIC_TIMESTAMP"])
	if err != nil {
		return JournalEntry{}, err
	}

	intPriority, err := strconv.Atoi(raw["PRIORITY"].(string))
	if err != nil {
		return JournalEntry{}, err
	}

	intSeqNum, err := strconv.Atoi(raw["__SEQNUM"].(string))
	if err != nil {
		return JournalEntry{}, err
	}

	intSyslogFacility, err := strconv.Atoi(raw["SYSLOG_FACILITY"].(string))
	if err != nil {
		return JournalEntry{}, err
	}

	event := JournalEntry{
		RealtimeTimestamp: realtimeTimestamp,
		MonotonicTimestamp: monotonicTimestamp,
		Message: raw["MESSAGE"].(string),
		Priority: intPriority,
		Cursor: raw["__CURSOR"].(string),
		SeqNum: intSeqNum,
		SeqNumId: raw["__SEQNUM_ID"].(string),
		SyslogFacility: intSyslogFacility,
		SyslogIdentifier: raw["SYSLOG_IDENTIFIER"].(string),
		Fields:    map[string]string{},
	}

	for key, value := range raw {
		event.Fields[key] = fmt.Sprint(value)
	}

	return event, nil
}

func parseTimestamp(v any) (time.Time, error) {
	value, ok := v.(string)
	if !ok {
		return time.Time{}, errors.New("invalid timestamp")
	}

	micros, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.UnixMicro(micros).UTC(), nil
}
