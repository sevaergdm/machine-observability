package gpu

import "time"

type Entry struct {
	BootId string    `parquet:"boot_id" json:"boot_id" doc:"boot UUID taken from /proc/sys/kernel/random/boot_id (cross-source contract column)"`
	Ts     time.Time `parquet:"ts,timestamp(microsecond)" json:"ts" doc:"timestamp in UTC when /proc/meminfo was polled (cross-source contract column)"`

	Card        string `parquet:"card" json:"card" doc:"drm card name (card1, card2); stable within a boot, NOT across boots (driver load order)"`
	BusyPercent int64  `parquet:"busy_percent" json:"busy_percent" doc:"%; gauge; fraction of time the GPU was executing over the driver's sampling window"`
	VramUsed    int64  `parquet:"vram_used" json:"vram_used" doc:"bytes; gauge; dedicated video memory in use"`
	VramTotal   int64  `parquet:"vram_total" json:"vram_total" doc:"bytes; gauge; dedicated video memory capacity"`
	GttUsed     int64  `parquet:"gtt_used" json:"gtt_used" doc:"bytes; gauge; system RAM mapped for GPU use (GTT) - spillover when VRAM is tight"`

	TempEdge     *int64 `parquet:"temp_edge" json:"temp_edge" doc:"°C (file is millidegrees ÷ 1000); gauge; die-edge sensor; null = sensor absent"`
	TempJunction *int64 `parquet:"temp_junction" json:"temp_junction" doc:"°C ÷ 1000; gauge; hottest spot sensor - the throttling trigger; null = absent (APU)"`
	TempMem      *int64 `parquet:"temp_mem" json:"temp_mem" doc:"°C ÷ 1000; gauge; VRAM temperature; null = absent (APU)"`
	PowerWatts   *int64 `parquet:"power_watts" json:"power_watts" doc:"W (file is microwatts ÷ 1e6); gauge; driver-smoothed board power draw; null = absent"`
	FanRpm       *int64 `parquet:"fan_rpm" json:"fan_rpm" doc:"RPM; gauge; null = no dedicated fan (APU shares chassis cooling)"`
	SclkMhz      *int64 `parquet:"sclk_mhz" json:"sclk_mhz" doc:"MHz (file is Hz ÷ 1e6); gauge; current core/shader clock; null = absent"`
	MclkMhz      *int64 `parquet:"mclk_mhz" json:"mclk_mhz" doc:"MHz ÷ 1e6; gauge; current memory clock; null = absent (APU)"`
}

func (e Entry) Source() string       { return "gpu" }
func (e Entry) Timestamp() time.Time { return e.Ts }
