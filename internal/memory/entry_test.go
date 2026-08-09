package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/parquet-go/parquet-go"
)

var entryA = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719260315896),
	Total:     int64(32094844) * 1024,
	Free:      int64(20195568) * 1024,
	Available: int64(27233332) * 1024,
	Buffers:   int64(2728) * 1024,
	Cached:    int64(7083384) * 1024,
	SwapTotal: int64(38203168) * 1024,
	SwapFree:  int64(38203168) * 1024,
	Dirty:     int64(3912) * 1024,
	Slab:      int64(775948) * 1024,
}

var entryB = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719270315896),
	Total:     int64(32094844) * 1024,
	Free:      int64(20195568) * 1024,
	Available: int64(27233332) * 1024,
	Buffers:   int64(2728) * 1024,
	Cached:    int64(7083384) * 1024,
	SwapTotal: int64(38203168) * 1024,
	SwapFree:  int64(38203168) * 1024,
	Dirty:     int64(3912) * 1024,
	Slab:      int64(775948) * 1024,
}

var entryC = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719280315896),
	Total:     int64(32094844) * 1024,
	Free:      int64(20195568) * 1024,
	Available: int64(27233332) * 1024,
	Buffers:   int64(2728) * 1024,
	Cached:    int64(7083384) * 1024,
	SwapTotal: int64(38203168) * 1024,
	SwapFree:  int64(38203168) * 1024,
	Dirty:     int64(3912) * 1024,
	Slab:      int64(775948) * 1024,
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
