package sink

import (
	"fmt"
	"io/fs"
	"log/slog"
	"machine-observability/internal/collector"
	"os"
	"path/filepath"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/parquet-go/parquet-go"
)

func NewParquetFlush[T any](dataDir, source string) FlushFunc {
	return func(rows []collector.Event) error {
		out := make([]T, 0, len(rows))
		for _, event := range rows {
			row, ok := event.(T)
			if !ok {
				return fmt.Errorf("flush for %q got event of type %T, not the expected row type", source, row)
			}
			out = append(out, row)
		}

		// Split partitions by the first row timestamp. It is possible that at hour boundaries some events will spill over, but as these
		// are partitions of event data, it should only be considered coincidental that they match the event time and not relied upon for
		// event occurrence filtering
		partition := rows[0].Timestamp().UTC()
		path := filepath.Join(dataDir, "source="+source, "date="+partition.Format("2006-01-02"), fmt.Sprintf("hour=%02d", partition.Hour()))
		err := os.MkdirAll(path, 0750)
		if err != nil {
			return fmt.Errorf("encountered an error creating directory %s: %v", path, err)
		}

		name := ulid.Make().String()
		tmp := filepath.Join(path, name+".parquet.tmp")
		final := filepath.Join(path, name+".parquet")

		f, err := os.Create(tmp)
		if err != nil {
			return fmt.Errorf("encountered an error creating parquet file: %v", err)
		}

		writer := parquet.NewGenericWriter[T](f)
		_, err = writer.Write(out)
		if err != nil {
			return fmt.Errorf("encountered an error writing to parquet file: %v", err)
		}
		err = writer.Close()
		if err != nil {
			return fmt.Errorf("encountered an error closing parquet writer: %v", err)
		}

		err = f.Sync()
		if err != nil {
			return fmt.Errorf("encountered an error syncing to disk: %v", err)
		}

		err = f.Close()
		if err != nil {
			return fmt.Errorf("encountered an error closing file: %v", err)
		}

		err = os.Rename(tmp, final)
		if err != nil {
			return fmt.Errorf("encountered an error renaming file '%s' to '%s': %v", tmp, final, err)
		}

		return nil
	}
}

func CleanTmp(dataDir string, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".parquet.tmp") {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		logger.Warn("removed leftover tmp file from unclean shutdown", "path", path)
		return nil
	})
}
