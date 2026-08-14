package gpu

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/parquet-go/parquet-go"
)

var entryA = Entry{
	BootId:       "boot-123",
	Ts:           time.UnixMicro(1784719280315896),
	VramUsed:     1000000,
	VramTotal:    32000000000,
	GttUsed:      8000000000,
	TempEdge:     new(int64(50)),
	TempJunction: new(int64(50)),
	TempMem:      new(int64(50)),
	PowerWatts:   new(int64(240)),
	FanRpm:       new(int64(2000)),
	SclkMhz:      new(int64(5)),
	MclkMhz:      new(int64(5)),
}

var entryB = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719280315896),
	VramUsed:  1000000,
	VramTotal: 32000000000,
	GttUsed:   8000000000,
}

var entryC = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719280315896),
	VramUsed:  1000000,
	VramTotal: 32000000000,
	GttUsed:   8000000000,
	SclkMhz:   new(int64(5)),
	MclkMhz:   new(int64(5)),
}

func TestEntryParquetRoundTrip(t *testing.T) {
	entries := []Entry{entryA, entryB, entryC}

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
