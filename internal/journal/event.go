package journal

import "time"

type Entry struct {
	Cursor             string    `json:"cursor"`
	MonotonicTimestamp time.Time `json:"monotonic_ts"`
	RealtimeTimestamp  time.Time `json:"realtime_ts"`
	SeqNum             int64     `json:"seqnum"`
	SeqNumId           string    `json:"seqnum_id"`

	Message          *string `json:"message"`
	Priority         *int64  `json:"priority"`
	SyslogFacility   *int64  `json:"syslog_facility"`
	SyslogIdentifier *string `json:"syslog_identifier"`

	SystemdUnit *string `json:"systemd_unit"`
	Pid         *int64  `json:"pid"`
	Uid         *int64  `json:"uid"`
	Comm        *string `json:"comm"`
	BootId      *string `json:"boot_id"`
	Transport   *string `json:"transport"`

	Fields string `json:"fields"`
}

func (e Entry) Source() string { return "journal" }

func (e Entry) Timestamp() time.Time { return e.RealtimeTimestamp }
