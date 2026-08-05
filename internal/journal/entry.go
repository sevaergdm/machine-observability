package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Cursor             string    `parquet:"cursor" json:"cursor"`
	MonotonicTimestamp int64     `parquet:"monotonic_ts" json:"monotonic_ts"`
	RealtimeTimestamp  time.Time `parquet:"realtime_ts,timestamp(microsecond)" json:"realtime_ts"`
	SeqNum             int64     `parquet:"seqnum" json:"seqnum"`
	SeqNumId           string    `parquet:"seqnum_id" json:"seqnum_id"`

	Message          *string `parquet:"message,optional" json:"message"`
	Priority         *int64  `parquet:"priority,optional" json:"priority"`
	SyslogFacility   *int64  `parquet:"syslog_facility,optional" json:"syslog_facility"`
	SyslogIdentifier *string `parquet:"syslog_identifier,optional" json:"syslog_identifier"`

	SystemdUnit *string `parquet:"systemd_unit,optional" json:"systemd_unit"`
	Pid         *int64  `parquet:"pid,optional" json:"pid"`
	Uid         *int64  `parquet:"uid,optional" json:"uid"`
	Comm        *string `parquet:"comm,optional" json:"comm"`
	BootId      *string `parquet:"boot_id,optional" json:"boot_id"`
	Transport   *string `parquet:"transport,optional" json:"transport"`

	Fields string `parquet:"fields" json:"fields"`
}

func (e Entry) Source() string { return "journal" }

func (e Entry) Timestamp() time.Time { return e.RealtimeTimestamp }

func (e Entry) WriteCursor(stateDir string) error {
	cursor := e.Cursor

	tmp := filepath.Join(stateDir, "journal.cursor.tmp")
	final := filepath.Join(stateDir, "journal.cursor")

	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("encountered an error creating cursor file: %w", err)
	}
	defer f.Close()

	if	_, err := f.WriteString(cursor); err != nil {
		return fmt.Errorf("encountered an error writing cursor to file: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("encountered an error syncing to disk: %w", err)
	}

	_ = f.Close()

	if err := os.Rename(tmp, final); err != nil {
		return fmt.Errorf("encountered an error renaming cursor file: %w", err)
	}

	return nil
}
