package journal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/parquet-go/parquet-go"
)

var fullEntry = Entry{
	Cursor: "s=065df9afa2ff4fe0a0b63d9c94a363ad;i=70b2d6;b=a7ee7b855a80493dadd29ca0054a802a;m=843e7d075;t=657315588f0f8;x=c010c66555fbf97",
	MonotonicTimestamp: 35499004021,
	RealtimeTimestamp: time.UnixMicro(1784719260315896),
	SeqNum: 7385814,
	SeqNumId: "065df9afa2ff4fe0a0b63d9c94a363ad",
	Message: new("gtk_widget_get_scale_factor: assertion 'GTK_IS_WIDGET (widget)' failed"),
	Priority: new(int64(4)),
	SyslogIdentifier: new("nm-applet"),
	SystemdUnit: new("user@1000.service"),
	Pid: new(int64(1311)),
	Uid: new(int64(1000)),
	Comm: new("nm-applet"),
	BootId: new("a7ee7b855a80493dadd29ca0054a802a"),
	Transport: new("journal"),
	Fields: `{"_SYSTEMD_INVOCATION_ID":"c3d7089fde054d59b6e782e281287cc9","GLIB_DOMAIN":"Gtk","_RUNTIME_SCOPE":"system","_AUDIT_LOGINUID":"1000","_HOSTNAME":"host","_CAP_EFFECTIVE":"0","_GID":"1000","GLIB_OLD_LOG_API":"1","_EXE":"/usr/bin/nm-applet","_SYSTEMD_USER_UNIT":"app-nm\\x2dapplet@autostart.service","_SOURCE_REALTIME_TIMESTAMP":"1784719260315825","_SYSTEMD_USER_SLICE":"app-graphical.slice","_SYSTEMD_OWNER_UID":"1000","_CMDLINE":"/usr/bin/nm-applet","_SYSTEMD_SLICE":"user-1000.slice","_MACHINE_ID":"0123456789abcdef0123456789abcdef","_AUDIT_SESSION":"3","_SYSTEMD_CGROUP":"/user.slice/user-1000.slice/user@1000.service/app.slice/app-graphical.slice/app-nm\\x2dapplet@autostart.service"}`, 
}

var minimalEntry = Entry{
	Cursor: "s=065df9afa2ff4fe0a0b63d9c94a363ad;i=70b2d6;b=a7ee7b855a80493dadd29ca0054a802a;m=843e7d075;t=657315588f0f8;x=c010c66555fbf97",
	MonotonicTimestamp: 35499004021,
	RealtimeTimestamp: time.UnixMicro(1784719260315896),
	SeqNum: 7385814,
	SeqNumId: "065df9afa2ff4fe0a0b63d9c94a363ad",
	Fields: `{}`,
}

var byteArrayEntry = Entry{
	Cursor: "s=065df9afa2ff4fe0a0b63d9c94a363ad;i=70b2d6;b=a7ee7b855a80493dadd29ca0054a802a;m=843e7d075;t=657315588f0f8;x=c010c66555fbf97",
	MonotonicTimestamp: 35499004021,
	RealtimeTimestamp: time.UnixMicro(1784719260315896),
	SeqNum: 7385814,
	SeqNumId: "065df9afa2ff4fe0a0b63d9c94a363ad",
	Message: new("binary test \uFFFD\uFFFD end"),
	Fields: `{}`,
}


func TestEntryParquetRoundTrip(t *testing.T) {
	entries := []Entry{fullEntry, minimalEntry, byteArrayEntry}	

	dir := t.TempDir()
	if keep := os.Getenv("PARQUET_OUT"); keep != "" {
		dir = keep
	}
	path := filepath.Join(dir, "entries.parquet")
	t.Log("wrote", path)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("error creating path: %v", err)
	}
	
	writer := parquet.NewGenericWriter[Entry](f)
	if _, err := writer.Write(entries); err != nil {
		t.Fatalf("error writing parquet file: %v", err)
	}
	writer.Close()
	f.Close()

	got, err := parquet.ReadFile[Entry](path)
	if err != nil {
		t.Fatalf("error reading parquet file: %v", err)
	}

	if diff := cmp.Diff(entries, got); diff != "" {
		t.Errorf("round-trip mismatch (-want, +got):\n%s", diff)
	}
}

