package memory

import "time"

type Entry struct {
	BootId string    `parquet:"boot_id" json:"boot_id" doc:"boot UUID taken from /proc/sys/kernel/random/boot_id (cross-source contract column)"`
	Ts     time.Time `parquet:"ts,timestamp(microsecond)" json:"ts" doc:"timestamp in UTC when /proc/meminfo was polled (cross-source contract column)"`

	Total     int64 `parquet:"total" json:"total" doc:"bytes; gauge; usable RAM (physical minus kernal-reserved); MemTotal; meminfo 'kB' = KiB"`
	Free      int64 `parquet:"free" json:"free" doc:"bytes; gauge; fully unused RAM. Low free is NORMAL (page cache eats spare RAM); MemFree; use available for pressure"`
	Available int64 `parquet:"available" json:"available" doc:"bytes; gauge; kernel estimate of memory claimable by new work without swapping; MemAvailable; the number to watch/alert on"`
	Buffers   int64 `parquet:"buffers" json:"buffers" doc:"bytes; gauge; block-device metadata cache; Buffers"`
	Cached    int64 `parquet:"cached" json:"cached" doc:"bytes; gauge; page cache (file contents held in RAM, reclaimable); Cached; excludes Buffers"`
	SwapTotal int64 `parquet:"swap_total" json:"swap_total" doc:"bytes; gauge; swap capacity; SwapTotal"`
	SwapFree  int64 `parquet:"swap_free" json:"swap_free" doc:"bytes; gauge; unused swap; SwapFree"`
	Dirty     int64 `parquet:"dirty" json:"dirty" doc:"bytes; gauge; page cache modified but not yet written back to disk; Dirty; sustained growth = writeback falling behind"`
	Slab      int64 `parquet:"slab" json:"slab" doc:"bytes; gauge; kernel object caches (inodes, dentries, ...); Slab"`
}

func (e Entry) Source() string       { return "memory" }
func (e Entry) Timestamp() time.Time { return e.Ts }
