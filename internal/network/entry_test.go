package network

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
	Ts:        time.UnixMicro(1784719280315896),
	Interface: "lo",
	RxBytes:   int64(100),
	RxPackets: int64(10),
	RxErrs:    int64(0),
	RxDrop:    int64(0),
	TxBytes:   int64(100),
	TxPackets: int64(10),
	TxErrs:    int64(0),
	TxDrop:    int64(0),
}

var entryB = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719280315896),
	Interface: "wlan0",
	RxBytes:   int64(100),
	RxPackets: int64(10),
	RxErrs:    int64(0),
	RxDrop:    int64(0),
	TxBytes:   int64(100),
	TxPackets: int64(10),
	TxErrs:    int64(0),
	TxDrop:    int64(0),
}

var entryC = Entry{
	BootId:    "boot-123",
	Ts:        time.UnixMicro(1784719280315896),
	Interface: "enp4s0",
	RxBytes:   int64(0),
	RxPackets: int64(0),
	RxErrs:    int64(0),
	RxDrop:    int64(0),
	TxBytes:   int64(0),
	TxPackets: int64(0),
	TxErrs:    int64(0),
	TxDrop:    int64(0),
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
