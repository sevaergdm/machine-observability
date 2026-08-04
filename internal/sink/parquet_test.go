package sink

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanTmp(t *testing.T) {
	dir := t.TempDir()

	partition := time.Now().UTC()
	path := filepath.Join(dir, "source=journal", "date="+partition.Format("2006-01-02"), fmt.Sprintf("hour=%02d", partition.Hour()))
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("unexpected error creating temp partition directory: %v", err)
	}

	tmpPath := filepath.Join(path, "x.parquet.tmp")
	finalPath := filepath.Join(path, "y.parquet")

	tmp, err := os.Create(tmpPath)
	if err != nil {
		t.Fatalf("unexpected error creating temp file: %v", err)
	}
	_ = tmp.Close()

	final, err := os.Create(finalPath)
	if err != nil {
		t.Fatalf("unexpected error creating final file: %v", err)
	}
	_ = final.Close()

	err = CleanTmp(dir, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(tmpPath); err == nil {
		t.Errorf("temp file %s was not deleted", tmpPath)
	}

	if _, err := os.Stat(finalPath); errors.Is(err, os.ErrNotExist) {
		t.Errorf("final file %s was deleted", finalPath)
	}
}
