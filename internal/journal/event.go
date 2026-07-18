package journal

import "time"

type JournalEntry struct {
	Cursor             string
	MonotonicTimestamp time.Time
	RealtimeTimestamp  time.Time
	SeqNum             int
	SeqNumId           string

	Message            *string
	Priority           *int
	SyslogFacility     *int
	SyslogIdentifier   *string

	Fields             map[string]string
}
