# pmx-tm

A CLI tool to monitor network traffic for each VM in a Proxmox cluster, storing
data in a local SQLite database.

It periodically reads network counters from Proxmox, computes the traffic
delta per VM, and stores it in SQLite. Query commands then let you read total,
daily, monthly, or range-based traffic — printed as a table or as JSON.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
  - [`start`](#start)
  - [`get`](#get)
  - [`get-daily`](#get-daily)
  - [`get-monthly`](#get-monthly)
  - [`get-range`](#get-range)
  - [`get-range-total`](#get-range-total)
  - [`clear`](#clear)
- [Global Flags](#global-flags)
- [Output Format](#output-format)
- [Configuration](#configuration)
- [Makefile Targets](#makefile-targets)

## Requirements

- Go 1.25+
- Access to a Proxmox VE cluster API (credentials configured via environment
  variables used by the bundled Proxmox client — see
  `pkg/api/proxmox`).

## Installation

Build for your current OS:

```sh
make build
```

Production build (Linux amd64, stripped):

```sh
make build-prod
```

Output is written to `./build/` as `pmx-tm`
(`.exe` on Windows).

## Usage

Running the binary with **no subcommand shows the help page**:

```sh
pmx-tm
```

Run a command:

```sh
pmx-tm <command> [id] [flags]
```

Where:

- `command` — one of the commands below.
- `id` — optional traffic record ID (`<node>-<vmid>-<name>`, e.g.
  `pve-101-web`). When omitted, the command targets all VMs.
- `flags` — command-specific options described below.

Every query command supports `--json` for machine-readable output.

## Commands

### `start`

Start the traffic monitor daemon. Collects network traffic deltas for all VMs
on a schedule and stores them in SQLite. Runs until interrupted (SIGINT/SIGTERM).

```
pmx-tm start [flags]
```

| Flag           | Shorthand | Description              |
| -------------- | --------- | ------------------------ |
| `--verbose`    | `-v`      | Enable info logging      |
| `--debug`      | `-d`      | Enable debug logging     |

A file lock (`app.pid` next to the config file) prevents running two instances
at the same time. If another instance is already running, `start` exits with an
error.

---

### `get`

Fetch total traffic for a specific VM or for all VMs (aggregated over all time).

```
pmx-tm get [id]
```

| Argument | Required | Description |
| -------- | -------- | ----------- |
| `id`     | no       | Traffic record ID. Omit for all VMs. |

Examples:

```sh
pmx-tm get pve-101-web
pmx-tm get
```

---

### `get-daily`

Fetch daily traffic for a specific VM or all VMs. The date defaults to today.

```
pmx-tm get-daily [id] [--date DD-MM-YYYY]
```

| Flag      | Required | Description                              |
| --------- | -------- | ---------------------------------------- |
| `--date`  | no       | Date in `DD-MM-YYYY` format (defaults to today) |
| `id`      | no       | Traffic record ID. Omit for all VMs.     |

Examples:

```sh
pmx-tm get-daily pve-101-web --date 10-08-2026
pmx-tm get-daily
```

---

### `get-monthly`

Fetch monthly traffic for a specific VM or all VMs. The month defaults to the
current month.

```
pmx-tm get-monthly [id] [--month YYYY-MM]
```

| Flag      | Required | Description                          |
| --------- | -------- | ------------------------------------ |
| `--month` | no       | Month in `YYYY-MM` format (defaults to current month) |
| `id`      | no       | Traffic record ID. Omit for all VMs. |

Examples:

```sh
pmx-tm get-monthly pve-101-web --month 2026-08
pmx-tm get-monthly
```

---

### `get-range`

Fetch traffic records for a specific VM or all VMs within a date range.

```
pmx-tm get-range [id] [--start DD-MM-YYYY] [--end DD-MM-YYYY]
```

| Flag      | Required | Description                            |
| --------- | -------- | -------------------------------------- |
| `--start` | no       | Start date in `DD-MM-YYYY` format      |
| `--end`   | no       | End date in `DD-MM-YYYY` format        |
| `id`      | no       | Traffic record ID. Omit for all VMs.   |

If omitted, the start defaults to the Unix epoch and the end to now. The end
date is inclusive (the full day is included).

Examples:

```sh
pmx-tm get-range pve-101-web --start 01-08-2026 --end 10-08-2026
pmx-tm get-range --start 01-08-2026 --end 10-08-2026
```

---

### `get-range-total`

Fetch the sum of all traffic (IN/OUT) for a specific VM or all VMs within a
date range.

```
pmx-tm get-range-total [id] [--start DD-MM-YYYY] [--end DD-MM-YYYY]
```

| Flag      | Required | Description                            |
| --------- | -------- | -------------------------------------- |
| `--start` | no       | Start date in `DD-MM-YYYY` format      |
| `--end`   | no       | End date in `DD-MM-YYYY` format        |
| `id`      | no       | Traffic record ID. Omit for all VMs.   |

If omitted, the start defaults to the Unix epoch and the end to now.

Examples:

```sh
pmx-tm get-range-total pve-101-web --start 01-08-2026 --end 10-08-2026
pmx-tm get-range-total
```

---

### `clear`

Clear traffic data for a specific VM or for all VMs.

```
pmx-tm clear [id] [--force]
```

| Flag      | Required | Description                                   |
| --------- | -------- | --------------------------------------------- |
| `--force` | no       | Skip the confirmation prompt                  |
| `id`      | no       | Traffic record ID. Omit to clear all VMs.     |

Without `--force`, prompts for confirmation (`y/n`). If no traffic data exists,
prints a notice and exits.

Examples:

```sh
pmx-tm clear pve-101-web --force
pmx-tm clear
```

---

## Global Flags

These flags apply to every command:

| Flag            | Shorthand | Description                                         |
| --------------- | --------- | --------------------------------------------------- |
| `--config`      | `-c`      | Path to configuration file (default `/var/lib/proxmox-traffic-monitor/config.yml`) |
| `--json`        |           | Output in JSON format                               |
| `--help`        | `-h`      | Show help                                           |

## Output Format

Human-readable output is a table:

```
ID            VMID  NODEID  DATE        IN        OUT
pve-101-web   101   pve     10 Aug 2026 1.2 GB    356 MB
```

- `get` and `get-range-total` hide the `DATE` column.
- Bytes are humanized (e.g. `1.2 GB`).

With `--json`, records are emitted as a JSON array. Each record:

```json
[
  {
    "id": "pve-101-web",
    "vmid": 101,
    "nodeid": "pve",
    "in": 1288490189,
    "out": 373293056,
    "timestamp": "2026-08-10T00:00:00Z"
  }
]
```

| Field       | Type           | Description                     |
| ----------- | -------------- | ------------------------------- |
| `id`        | string         | Traffic record ID               |
| `vmid`      | int64          | Proxmox VM ID                   |
| `nodeid`    | string         | Proxmox node ID                 |
| `in`        | uint64         | Incoming bytes                  |
| `out`       | uint64         | Outgoing bytes                  |
| `timestamp` | timestamp/null | Record time (omitted when nil, e.g. totals) |

## Configuration

A config file is created automatically at the `--config` path if it does not
exist. Defaults are applied for missing fields.

```yaml
app:
  sync_queue_workers: 4
  sync_queue_buffer: 100
  database: "/var/lib/proxmox-traffic-monitor/data.db"
  retention-period: 30
  fetch-interval: 30
```

| Field                | Default                                      | Description                     |
| -------------------- | -------------------------------------------- | ------------------------------- |
| `sync_queue_workers` | `4`                                          | Concurrent collection workers   |
| `sync_queue_buffer`  | `100`                                        | Job queue buffer size           |
| `database`           | `/var/lib/proxmox-traffic-monitor/data.db`   | Path to SQLite database         |
| `retention-period`   | `30`                                         | Days of history to keep (daily cleanup) |
| `fetch-interval`     | `30`                                         | Traffic collection interval in seconds |

The Proxmox API credentials are read from environment variables used by
`pkg/api/proxmox` (set `PVE_HOST`/host, user, and token/password accordingly
for your environment).

## Makefile Targets

| Target        | Description                              |
| ------------- | ---------------------------------------- |
| `make help`   | Show all Makefile targets                |
| `make clean`  | Remove all files from `build/`           |
| `make build`  | Build for the current OS                 |
| `make build-prod` | Build stripped Linux amd64 binary    |
| `make start`  | Run the built binary with `start`        |
