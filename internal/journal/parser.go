package journal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var promotedKeys = map[string]bool{
	"__REALTIME_TIMESTAMP":  true,
	"__MONOTONIC_TIMESTAMP": true,
	"MESSAGE":               true,
	"PRIORITY":              true,
	"__CURSOR":              true,
	"__SEQNUM":              true,
	"__SEQNUM_ID":           true,
	"SYSLOG_FACILITY":       true,
	"SYSLOG_IDENTIFIER":     true,
	"_SYSTEMD_UNIT":         true,
	"_PID":                  true,
	"_UID":                  true,
	"_COMM":                 true,
	"_BOOT_ID":              true,
	"_TRANSPORT":            true,
}

func Parse(raw map[string]any) (Entry, error) {
	realtimeTimestamp, err := parseTimestamp(raw["__REALTIME_TIMESTAMP"])
	if err != nil {
		return Entry{}, err
	}

	monotonicTimestamp, err := parseTimestamp(raw["__MONOTONIC_TIMESTAMP"])
	if err != nil {
		return Entry{}, err
	}

	seqNum := getInt(raw, "__SEQNUM")
	if seqNum == nil {
		return Entry{}, fmt.Errorf("missing or invalid __SEQNUM")
	}

	seqNumId := getString(raw, "__SEQNUM_ID")
	if seqNumId == nil {
		return Entry{}, fmt.Errorf("missing or invalid __SEQNUM_ID")
	}

	cursor := getString(raw, "__CURSOR")
	if cursor == nil {
		return Entry{}, fmt.Errorf("missing or invalid __CURSOR")
	}

	event := Entry{
		RealtimeTimestamp:  realtimeTimestamp,
		MonotonicTimestamp: monotonicTimestamp,
		Message:            getMessage(raw),
		Priority:           getInt(raw, "PRIORITY"),
		Cursor:             *cursor,
		SeqNum:             *seqNum,
		SeqNumId:           *seqNumId,
		SyslogFacility:     getInt(raw, "SYSLOG_FACILITY"),
		SyslogIdentifier:   getString(raw, "SYSLOG_IDENTIFIER"),
		SystemdUnit:        getString(raw, "_SYSTEMD_UNIT"),
		Pid:                getInt(raw, "_PID"),
		Uid:                getInt(raw, "_UID"),
		Comm:               getString(raw, "_COMM"),
		BootId:             getString(raw, "_BOOT_ID"),
		Transport:          getString(raw, "_TRANSPORT"),
		Fields:             "",
	}

	extra := make(map[string]any)
	for key, value := range raw {
		if promotedKeys[key] {
			continue
		}
		extra[key] = value
	}

	fieldsJSON, err := json.Marshal(extra)
	if err != nil {
		return Entry{}, fmt.Errorf("encoding extra fields: %w", err)
	}

	event.Fields = string(fieldsJSON)

	return event, nil
}

func getMessage(raw map[string]any) *string {
	if msgString := getString(raw, "MESSAGE"); msgString != nil {
		return msgString
	}

	msgArr, ok := raw["MESSAGE"].([]any)
	if !ok {
		return nil
	}

	b := make([]byte, 0, len(msgArr))
	for _, v := range msgArr {
		f, ok := v.(float64)
		if !ok {
			return nil
		}
		b = append(b, byte(f))
	}

	msgString := strings.ToValidUTF8(string(b), "\uFFFD")
	return &msgString
}

func getString(raw map[string]any, key string) *string {
	value, ok := raw[key].(string)
	if !ok {
		return nil
	}
	return &value
}

func getInt(raw map[string]any, key string) *int64 {
	value, ok := raw[key].(string)
	if !ok {
		return nil
	}

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}

	return &valueInt
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
