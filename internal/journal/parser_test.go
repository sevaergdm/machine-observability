package journal

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

//func ptr[T any](v T) *T { return new(v) }

func decode(t *testing.T, line string) map[string]any {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		t.Fatalf("bad test fixture: %v", err)
	}
	return raw
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Entry
		wantErr bool
	}{
		{
			name: "normal service entry",
			line: `{"__CURSOR":"s=abc;i=1f4","__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"5000000","__SEQNUM":"500","__SEQNUM_ID":"seq-1","PRIORITY":"6","SYSLOG_FACILITY":"3","SYSLOG_IDENTIFIER":"sshd","MESSAGE":"Accepted publickey for michael"}`,
			want: Entry{
				Cursor:             "s=abc;i=1f4",
				RealtimeTimestamp:  time.UnixMicro(1753142400000000).UTC(),
				MonotonicTimestamp: time.UnixMicro(5000000).UTC(),
				SeqNum:             500,
				SeqNumId:           "seq-1",
				Priority:           new(int64(6)),
				SyslogFacility:     new(int64(3)),
				SyslogIdentifier:   new("sshd"),
				Message:            new("Accepted publickey for michael"),
			},
		},
		{
			name: "minimal entry has nulls, not errors",
			line: `{"__CURSOR":"s=abc;i=1f4","__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"5000000","__SEQNUM":"500","__SEQNUM_ID":"seq-1"}`,
			want: Entry{
				Cursor:             "s=abc;i=1f4",
				RealtimeTimestamp:  time.UnixMicro(1753142400000000).UTC(),
				MonotonicTimestamp: time.UnixMicro(5000000).UTC(),
				SeqNum:             500,
				SeqNumId:           "seq-1",
			},
		},
		{
			name: "byte-array MESSAGE is decoded and sanitized",
			line: `{"__CURSOR":"s=abc;i=1f4","__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"5000000","__SEQNUM":"500","__SEQNUM_ID":"seq-1", "MESSAGE":[104,105,32,255]}`,
			want: Entry{
				Cursor:             "s=abc;i=1f4",
				RealtimeTimestamp:  time.UnixMicro(1753142400000000).UTC(),
				MonotonicTimestamp: time.UnixMicro(5000000).UTC(),
				SeqNum:             500,
				SeqNumId:           "seq-1",
				Message:            new("hi \uFFFD"), // 104,105,32 = "hi", 255 = invalid UTF-8
			},
		},
		{
			name:    "missing cursor is a parse failure",
			line:    `{"__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"5000000","__SEQNUM":"500","__SEQNUM_ID":"seq-1"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(decode(t, tt.line))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got.Fields = "{}"
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse mismatch\n got: %+v\nwant: %+v", got, tt.want)
				if got.Message != nil && tt.want.Message != nil {
					t.Errorf("message: got %q, want %q", *got.Message, *tt.want.Message)
				}
			}
		})
	}
}
