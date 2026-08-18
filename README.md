# Machine Observability

A single-binary observability agent for an Arch Linux machine. It collects system telemetry from journald logs, CPU, memory, disk, network, and AMD GPU. It writes the data as Hive-partitioned, ZSTD-compressed Parquet files that DuckDB (or anything else that speaks Parquet) can query in place.

This is a personal experimentation and learning project in Go. It is built for one machine (mine), but everything except the amdgpu collector is plain `/proc` parsing and should work on most Linux systems.

## What it collects

| Source | Origin | Kind | Notes |
|---|---|---|---|
| `journal` | `journalctl -f` (streaming) | events | cursor checkpoints: survives restarts without gaps (at-least-once delivery) |
| `cpu` | `/proc/stat` | cumulative counters | one row per CPU per tick, plus an `all` aggregate row |
| `memory` | `/proc/meminfo` | gauges | bytes; one row per tick |
| `disk_io` | `/proc/diskstats` | cumulative counters | whole devices only (from `/sys/block`), pseudo-devices excluded |
| `disk_fs` | `/proc/self/mountinfo` + `statfs(2)` | gauges | one row per real mount; size/used bytes + inodes |
| `network` | `/proc/net/dev` | cumulative counters | one row per interface per tick, `lo` included |
| `gpu` | amdgpu sysfs | gauges | one row per card per tick; sensors vary per card (nullable columns); a runtime-suspended card simply produces no rows |

### Design rules that apply everywhere

- Raw counters are stored as-is: rates and utilization are computed at query timewith window functions
- Every row carries `ts` (UTC, microsecond) and `boot_id`, so queries can handle counter resets across reboots
- Wide, typed tables per source: no generic key/value metrics table. Journal's long tail of fields lands in a single JSON column, queryable with DuckDB's `->>` operators
- Writes are atomic (tmp + fsync + rename); a crash never leaves a half-written Parquet file visible to queries

## Data layout

```
<data_dir>/
    source=journal/date=2026-08-18/hour=14/01J….parquet
    source=cpu/…
    source=disk_io/… source=disk_fs/… …
```

Hive-style partitioning; each flush is one ULID-named file. Buffers flush on row count, age, or shutdown.

## Install

```sh
git clone https://github.com/sevaergdm/machine-observability
cd machine-observability/packaging
makepkg -si
```

This builds from the latest release tag, runs the test suite (a failing test aborts the install), and installs:

- `/usr/bin/machine-observability`
- `/etc/machine-observability/config.toml` (marked as config: your edits survice upgrades)
- systemd unit + sysusers entry for the `machineobs` service user

Then:

```sh
sudo systemctl enable --now machine-observability
```

The service runs as `machineobs` (member of `systemd-journal`, so it can read the full journal) and writes under `/var/lib/machine-observability`. To query the data as your own user, add yourself to the `machineobs` group.

## Configure

`/etc/machine-observability/config.toml`:

```toml
data_dir  = "/var/lib/machine-observability/data"
state_dir = "/var/lib/machine-observability/state" 
log_level = "info"          # debug|info|warn|error

[collectors.journal]
enabled = true              # streaming: takes no interval

[collectors.cpu]
enabled = true
interval = "10s"

[collectors.memory]
enabled = true
interval = "10s"

[collectors.disk]           # emits both disk_io and disk_fs
enabled = true
interval = "60s"

[collectors.network]
enabled = true
interval = "10s"

[collectors.gpu]
enabled = true
interval = "10s"
```

Unknown collector names and misspelled keys are startup errors

## Querying

```sql
-- rows per source
SELECT source, COUNT(*)
FROM read_parquet('/var/lib/machine-observability/data/source=*/**/*.parquet', hive_partitioning=true, union_by_name=true)
GROUP BY source
;

-- noisiest journal units in the last day
SELECT systemd_unit, COUNT(*) AS n
FROM read_parquet('/var/lib/machine-observability/data/source=journal/**/*.parquet', hive_partitioning=true)
WHERE ts > now() - INTERVAL 1 DAY
GROUP BY 1
ORDER BY n DESC
LIMIT 20
;

-- CPU utilization from raw jiffies
SELECT
      ts
    , cpu
    , 100.0 * (1 - (idle - LAG(idle) OVER w) / CASE((user + nice + system + idle + iowait + irq + softirq + steal) OVER w AS DOUBLE)) AS util_pct
FROM read_parquet('/var/lib/machine-observability/data/source=cpu/**/*.parquet', hive_partitioning=true)
WHERE cpu = 'all'
WINDOW w AS (PARTITION BY boot_id, cpu ORDER BY ts)
;
```

Mind timezones when comparing against `journalctl`: Parquet timestamps are UTC; use offset-explict literals (`'… 14:00:00+02:00'`).

## Development

```sh
make build                          # -> bin/agent
make test                           # all packages
make vet fmt-check
./bin/agent -config config.toml     # foreground run against a local config
```

Every parser is a pure function over an `io.Reader` with fixtures from a real machine under `testdata/`; collector machinery is tested against fakes. `go test -race ./...` is expected to pass.

Volume, for scale: on the author's (healthy) machine the full collector set produces ~180k rows -> ~18 MB/day compressed. It once measured a PCIe interrupt storm at 28 events/sec that a broken power fix was spraying into the journal, which is exactly the kind of thing it exists to catch.

## License

MIT
