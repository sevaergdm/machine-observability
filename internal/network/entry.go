package network

import "time"

type Entry struct {
	BootId string    `parquet:"boot_id" json:"boot_id" doc:"boot UUID taken from /proc/sys/kernel/random/boot_id (cross-source contract column)"`
	Ts     time.Time `parquet:"ts,timestamp(microsecond)" json:"ts" doc:"timestamp in UTC when /proc/net/dev was polled (cross-source contract column)"`

	Interface string `parquet:"interface" json:"interface" doc:"interface name; row key; includes lo (loopback deliberately collected)"`
	RxBytes   int64  `parquet:"rx_bytes" json:"rx_bytes" doc:"bytes received; cumulative counter; /proc/net/dev rx value 1"`
	RxPackets int64  `parquet:"rx_packets" json:"rx_packets" doc:"packets received; cumulative counter; /proc/net/dev rx value 2"`
	RxErrs    int64  `parquet:"rx_errs" json:"rx_errs" doc:"malformed/failed packets on receive (CRC, framing - link-layer trouble); cumulative counter; rx value 3"`
	RxDrop    int64  `parquet:"rx_drop" json:"rx_drop" doc:"intact packets discarded on receive (buffer/memory exhaustion - overload, not breakage); cumulative counter; rx value 4"`
	TxBytes   int64  `parquet:"tx_bytes" json:"tx_bytes" doc:"bytes transmitted; cumulative counter; /proc/net/dev tx value 1"`
	TxPackets int64  `parquet:"tx_packets" json:"tx_packets" doc:"packets transmitted; cumulative counter; /proc/net/dev tx value 2"`
	TxErrs    int64  `parquet:"tx_errs" json:"tx_errs" doc:"transmit failures (carrier loss, collisions); cumulative counter; tx value 3"`
	TxDrop    int64  `parquet:"tx_drop" json:"tx_drop" doc:"packets discarded before transmit (queue full - overload); cumulative counter; tx value 4"`
}

func (e Entry) Source() string       { return "network" }
func (e Entry) Timestamp() time.Time { return e.Ts }
