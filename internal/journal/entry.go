package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Cursor             string    `parquet:"cursor" json:"cursor" doc:"opaque journald position token, unique per entry; __CURSOR; checkpoint + dedupkey"`
	MonotonicTimestamp int64     `parquet:"monotonic_ts" json:"monotonic_ts" doc:"microseconds since boot (monotonic clock, NOT wall time); __MONOTONIC_TIMESTAMP"`
	Ts                 time.Time `parquet:"ts,timestamp(microsecond)" json:"ts" doc:"UTC wall-clock time journald received the entry; __REALTIME_TIMESTAMP; event time  (cross-source contract column)"`
	SeqNum             int64     `parquet:"seqnum" json:"seqnum" doc:"journald per-series sequence number; __SEQNUM"`
	SeqNumId           string    `parquet:"seqnum_id" json:"seqnum_id" doc:"id of the seqnum series; __SEQNUM_ID"`

	Message          *string `parquet:"message,optional" json:"message" doc:"log text, sanitized to valid UTF-8 (binary payloads lossy via U+FFFD); null = entry had no MESSAGE"`
	Priority         *int64  `parquet:"priority,optional" json:"priority" doc:"syslog severity 0(emerg)-7(debug); null on audit/some kernel entries"`
	SyslogFacility   *int64  `parquet:"syslog_facility,optional" json:"syslog_facility" doc:"syslog facility code (0=kern, 3=daemon, ...); null when not syslog-originated"`
	SyslogIdentifier *string `parquet:"syslog_identifier,optional" json:"syslog_identifier" doc:"sender-CLAIMED tag (process-supplied, not verified); null when absent"`

	SystemdUnit *string `parquet:"systemd_unit,optional" json:"systemd_unit" doc:"originating unit, journald-verified; _SYSTEMD_UNIT; null = not from a unit (e.g. kernel)"`
	Pid         *int64  `parquet:"pid,optional" json:"pid" doc:"sender pid, journald-verified; _PID; null when no process (kernel)"`
	Uid         *int64  `parquet:"uid,optional" json:"uid" doc:"sender uid, journald-verified; _UID; null when no process"`
	Comm        *string `parquet:"comm,optional" json:"comm" doc:"process command name, journald-verified, kernel-truncated to 16 chars; __COMM"`
	BootId      *string `parquet:"boot_id,optional" json:"boot_id" doc:"boot UUID; _BOOT_ID (cross-source contract column)"`
	Transport   *string `parquet:"transport,optional" json:"transport" doc:"path into journald; kernel|syslog|journal|stdout|audit; _TRANSPORT"`

	Fields string `parquet:"fields" json:"fields" doc:"all remaining journald fields as a JSON object string; promoted keys excluded; '{}' when none; query via fields->>'_KEY'"`
}

func (e Entry) Source() string       { return "journal" }
func (e Entry) Timestamp() time.Time { return e.Ts }

func (e Entry) WriteCursor(stateDir string) error {
	cursor := e.Cursor

	tmp := filepath.Join(stateDir, "journal.cursor.tmp")
	final := filepath.Join(stateDir, "journal.cursor")

	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("encountered an error creating cursor file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(cursor); err != nil {
		return fmt.Errorf("encountered an error writing cursor to file: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("encountered an error syncing to disk: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("encountered an error closing cursor file: %w", err)
	}

	if err := os.Rename(tmp, final); err != nil {
		return fmt.Errorf("encountered an error renaming cursor file: %w", err)
	}

	return nil
}
