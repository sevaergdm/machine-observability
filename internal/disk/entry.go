package disk

import "time"

type IOEntry struct {
	BootId string    `parquet:"boot_id" json:"boot_id" doc:"boot UUID taken from /proc/sys/kernel/random/boot_id (cross-source contract column)"`
	Ts     time.Time `parquet:"ts,timestamp(microsecond)" json:"ts" doc:"timestamp in UTC when /proc/diskstats was polled (cross-source contract column)"`

	DeviceName      string `parquet:"device_name" json:"device_name" doc:"device name from /proc/diskstats. Matched against whole-device set in /sys/block. Partitions excluded"`
	ReadsCompleted  int64  `parquet:"reads_completed" json:"reads_completed" doc:"operations; cumulative counter; number of reads completed successfully"`
	ReadsMerged     int64  `parquet:"reads_merged" json:"reads_merged" doc:"operations; cumulative counter; number of reads merged (adjacent reads and writes may be merged)"`
	ReadSectors     int64  `parquet:"read_sectors" json:"read_sectors" doc:"sector = 512 bytes; cumulative counter; number of sectors read successfully"`
	ReadMs          int64  `parquet:"read_ms" json:"read_ms" doc:"milliseconds; cumulative counter; total number of milliseconds spent reading"`
	WritesCompleted int64  `parquet:"writes_completed" json:"writes_completed" doc:"operations; cumulative counter; number of writes completed successfully"`
	WritesMerged    int64  `parquet:"writes_merged" json:"writes_merged" doc:"operations; cumulative counter; number of writes merged (adjacent reads and writes may be merged)"`
	WriteSectors    int64  `parquet:"write_sectors" json:"write_sectors" doc:"sector = 512 bytes; cumulative counter; number of sectors written successfully"`
	WriteMs         int64  `parquet:"write_ms" json:"write_ms" doc:"milliseconds; cumulative counter; total number of milliseconds spent writing"`
	IoInProgress    int64  `parquet:"io_in_progress" json:"io_in_progress" doc:"operations; gauge; total number of i/o operations currently in progress"`
	IoMs            int64  `parquet:"io_ms" json:"io_ms" doc:"milliseconds; cumulative counter; total number of milliseconds doing i/o operations"`
	IoMsWeighted    int64  `parquet:"io_ms_weighted" json:"io_ms_weighted" doc:"request-milliseconds: each ms adds the count of I/Os then in flight (integral of queue depth over time); cumulative; Δ/Δt = avg queue depth, Δ/Δops = avg request latency including queueing"`
}

func (e IOEntry) Source() string       { return "disk_io" }
func (e IOEntry) Timestamp() time.Time { return e.Ts }

type FSEntry struct {
	BootId string    `parquet:"boot_id" json:"boot_id" doc:"boot UUID taken from /proc/sys/kernel/random/boot_id (cross-source contract column)"`
	Ts     time.Time `parquet:"ts,timestamp(microsecond)" json:"ts" doc:"timestamp in UTC when /proc/self/mountinfo was polled (cross-source contract column)"`

	Mount  string `parquet:"mount" json:"mount" doc:"where the file system is mounted, e.g. '/', '/home', etc."`
	FsType string `parquet:"fs_type" json:"fs_type" doc:"the kind of file system: 'btrfs', 'ext4', 'vfat'"`
	Device string `parquet:"device" json:"device" doc:"which device backs the file system, e.g. '/dev/nvme0n1p2'"`

	SizeBytes  int64 `parquet:"size_bytes" json:"size_bytes" doc:"total capacity; gauge"`
	UsedBytes  int64 `parquet:"used_bytes" json:"used_bytes" doc:"consumed capacity; gauge"`
	FilesTotal int64 `parquet:"files_total" json:"files_total" doc:"inode capacity; gauge"`
	FilesUsed  int64 `parquet:"files_used" json:"files_used" doc:"consumed inode capacity; gauge"`
}

func (e FSEntry) Source() string       { return "disk_fs" }
func (e FSEntry) Timestamp() time.Time { return e.Ts }
