package journal

import "time"

type JournalEntry struct {
	Cursor             string
	MonotonicTimestamp time.Time
	RealtimeTimestamp  time.Time
	Message            string
	Priority           int
	SeqNum             int
	SeqNumId           string
	SyslogFacility     int
	SyslogIdentifier   string
	Fields             map[string]string
}
