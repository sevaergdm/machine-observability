package cpu

import "time"

type Entry struct {
	BootId string    `parquet:"boot_id" json:"boot_id"`
	Ts     time.Time `parquet:"ts" json:"ts"`

	Cpu     string `parquet:"cpu" json:"cpu" doc:"cpu being measured, where 'all' is the aggregation of all cpus"`
	User    int64  `parquet:"user" json:"user" doc:"jiffies (1/100s) in user space; cumulative since boot; /proc/stat value 1"`
	Nice    int64  `parquet:"nice" json:"nice" doc:"jiffies (1/100s) in low-priority (niced) user space; cumulative since boot; /proc/stat value 2"`
	System  int64  `parquet:"system" json:"system" doc:"jiffies (1/100s) of kernel time (syscalls, drivers); cumulative since boot; /proc/stat value 3"`
	Idle    int64  `parquet:"idle" json:"idle" doc:"jiffies (1/100s) doing nothing; cumulative since boot; /proc/stat value 4"`
	Iowait  int64  `parquet:"iowait" json:"iowait" doc:"jiffies (1/100s) idle while a disk I/O was pending; cumulative since boot; /proc/stat value 5"`
	Irq     int64  `parquet:"irq" json:"irq" doc:"jiffies (1/100s) servicing hardware interrupts; cumulative since boot; /proc/stat value 6"`
	SoftIrq int64  `parquet:"soft_irq" json:"soft_irq" doc:"jiffies (1/100s) servicing software interrupts; cumulative since boot; /proc/stat value 7"`
	Steal   int64  `parquet:"steal" json:"steal" doc:"jiffies (1/100s) a hypervisor stole time; cumulative since boot; /proc/stat value 8"`
}

func (e Entry) Source() string { return "cpu" }

func (e Entry) Timestamp() time.Time { return e.Ts }
