package cpu

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/parquet-go/parquet-go"
)

var EntryA = Entry{
	BootId:  "boot-123",
	Ts:      time.UnixMicro(1784719260315896),
	Cpu:     "all",
	User:    200,
	Nice:    200,
	System:  200,
	Idle:    200,
	Iowait:  200,
	Irq:     200,
	SoftIrq: 200,
	Steal:   200,
}

var EntryB = Entry{
	BootId:  "boot-123",
	Ts:      time.UnixMicro(1784719260315896),
	Cpu:     "0",
	User:    100,
	Nice:    100,
	System:  100,
	Idle:    100,
	Iowait:  100,
	Irq:     100,
	SoftIrq: 100,
	Steal:   100,
}

var EntryC = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719260315896),
	Cpu:     "1",
	User:    100,
	Nice:    100,
	System:  100,
	Idle:    100,
	Iowait:  100,
	Irq:     100,
	SoftIrq: 100,
	Steal:   100,
}

func TestEntryParquetRoundTrip(t *testing.T) {
	entries := []Entry{EntryA, EntryB, EntryC}

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

	if err := writer.Close(); err != nil {
		t.Fatalf("error while closing writer: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("error while closing file: %v", err)
	}

	got, err := parquet.ReadFile[Entry](path)
	if err != nil {
		t.Fatalf("error reading parquet file: %v", err)
	}

	if diff := cmp.Diff(entries, got); diff != "" {
		t.Errorf("round-trip mismatch (-want, +got):\n%s", diff)
	}
}
