package journal

import "time"

type Entry struct {
	Cursor             string    `parquet:"cursor" json:"cursor"`
	MonotonicTimestamp int64     `parquet:"monotonic_ts,int64(microsecond)" json:"monotonic_ts"`
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
