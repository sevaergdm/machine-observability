package disk

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/parquet-go/parquet-go"
)

var entryIOA = IOEntry{
	BootId:          "boot-123",
	Ts:              time.UnixMicro(1784719260315896),
	DeviceName:      "nvme0n1",
	ReadsCompleted:  100,
	ReadsMerged:     100,
	ReadSectors:     100,
	ReadMs:          10000,
	WritesCompleted: 100,
	WritesMerged:    100,
	WriteSectors:    100,
	WriteMs:         10000,
	IoInProgress:    0,
	IoMs:            10000,
	IoMsWeighted:    1000,
}

var entryIOB = IOEntry{
	BootId:          "boot-123",
	Ts:              time.UnixMicro(1784719260315896),
	DeviceName:      "nvme0n1",
	ReadsCompleted:  100,
	ReadsMerged:     100,
	ReadSectors:     100,
	ReadMs:          10000,
	WritesCompleted: 100,
	WritesMerged:    100,
	WriteSectors:    100,
	WriteMs:         10000,
	IoInProgress:    0,
	IoMs:            10000,
	IoMsWeighted:    1000,
}

var entryIOC = IOEntry{
	BootId:          "boot-123",
	Ts:              time.UnixMicro(1784719260315896),
	DeviceName:      "nvme0n1",
	ReadsCompleted:  100,
	ReadsMerged:     100,
	ReadSectors:     100,
	ReadMs:          10000,
	WritesCompleted: 100,
	WritesMerged:    100,
	WriteSectors:    100,
	WriteMs:         10000,
	IoInProgress:    0,
	IoMs:            10000,
	IoMsWeighted:    1000,
}

func TestIOEntryParquetRoundTrip(t *testing.T) {
	entries := []IOEntry{entryIOA, entryIOB, entryIOC}

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

	writer := parquet.NewGenericWriter[IOEntry](f)
	if _, err := writer.Write(entries); err != nil {
		t.Fatalf("error writing parquet file: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("error while closing writer: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("error while closing file: %v", err)
	}

	got, err := parquet.ReadFile[IOEntry](path)
	if err != nil {
		t.Fatalf("error reading parquet file: %v", err)
	}

	if diff := cmp.Diff(entries, got); diff != "" {
		t.Errorf("round-trip mismatch (-want, +got):\n%s", diff)
	}
}
