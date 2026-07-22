package journal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
)

func decode(t *testing.T, line string) map[string]any {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		t.Fatalf("bad test fixture: %v", err)
	}
	return raw
}

func TestParseSyntheticData(t *testing.T) {
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

			got.Fields = ""
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Errorf("Parse mismatch: %v", diff)
			}
		})
	}
}

func TestParseFields(t *testing.T) {
	line := `{"__CURSOR":"s=abc;i=1f4","__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"5000000","__SEQNUM":"500","__SEQNUM_ID":"seq-1", "MESSAGE":"hello","_EXE":"/usr/bin/sshd","_CMDLINE":"sshd: michael [priv]"}`

	got, err := Parse(decode(t, line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal([]byte(got.Fields), &fields); err != nil {
		t.Fatalf("fields is not valid json: %v", err)
	}

	if exe, _ := fields["_EXE"].(string); exe != "/usr/bin/sshd" {
		t.Errorf("_EXE = %q, want: %q", exe, "/usr/bin/sshd")
	}

	for _, key := range []string{"MESSAGE", "__CURSOR", "PRIORITY", "_PID"} {
		if _, present := fields[key]; present {
			t.Errorf("promoted key %s should not be in fields", key)
		}
	}
}

func TestParseFieldsEmpty(t *testing.T) {
	line := `{"__CURSOR":"s=abc;i=1f4","__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"5000000","__SEQNUM":"500","__SEQNUM_ID":"seq-1", "MESSAGE":"hello","_PID":"1234"}`

	got, err := Parse(decode(t, line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Fields != "{}" {
		t.Errorf("expected empty object, but got: %q", got.Fields)
	}
}

func TestParseRealFixtures(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
			if err != nil {
				t.Fatalf("reading fixture: %v", err)
			}

			got, err := Parse(decode(t, string(data)))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Cursor == "" {
				t.Error("cursor is empty")
			}

			if got.Message != nil && !utf8.ValidString(*got.Message) {
				t.Error("message is not valid UTF-8")
			}

			var Fields map[string]any
			if err := json.Unmarshal([]byte(got.Fields), &Fields); err != nil {
				t.Errorf("fields is not valid JSON: %v", err)
			}
		})
	}
}
