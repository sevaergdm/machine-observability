package journal

import "time"

type Entry struct {
	Cursor             string
	MonotonicTimestamp time.Time
	RealtimeTimestamp  time.Time
	SeqNum             int
	SeqNumId           string

	Message          *string
	Priority         *int
	SyslogFacility   *int
	SyslogIdentifier *string

	Fields map[string]string
}

func (e Entry) Source() string { return "journal" }

func (e Entry) Timestamp() time.Time { return e.RealtimeTimestamp }
