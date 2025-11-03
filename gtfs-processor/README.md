# TITSA GTFS Data Processor

A high-performance CLI tool to process GTFS (General Transit Feed Specification) data for TITSA bus schedules, merge multiple CSV files, and generate optimized Parquet output.

## Features

- 🚀 **Fast**: Efficient memory usage for processing large datasets
- 📊 **Parquet Output**: Compressed columnar format for optimal storage and querying
- 🎨 **Beautiful CLI**: Clean, colorful interface with progress indicators
- ⏰ **Smart Time Handling**: Properly handles GTFS times (24:00-29:59 for next-day arrivals)
- 🔍 **Flexible Filtering**: Filter by stop ID and date range

## Installation

### Prerequisites

- Go 1.21 or later

### Build from Source

```bash
cd gtfs-processor
go mod download
go build -o gtfs-processor
```

### Cross-compile for Linux (from macOS)

```bash
GOOS=linux GOARCH=amd64 go build -o gtfs-processor-linux
```

## Usage

The CLI has three main commands:

### 1. Process Only

Process GTFS files and generate Parquet output:

```bash
# Process all data (no filters)
./gtfs-processor process

# Filter by stop ID
./gtfs-processor process --stop-id 1918

# Filter by date range
./gtfs-processor process --start-date 20221018 --end-date 20221020

# Combine filters
./gtfs-processor process \
  --stop-id 1918 \
  --start-date 20221018 \
  --end-date 20221020 \
  --input ../batch_job/Google_transit \
  --output output/titsa_schedules.parquet
```

### 2. Upload Only

Upload an existing Parquet file to Tinybird:

```bash
# Upload with default settings
export TINYBIRD_TOKEN="your_token_here"
./gtfs-processor upload output/titsa_schedules.parquet

# Or specify token as flag
./gtfs-processor upload \
  --token "your_token" \
  --datasource titsa_schedules \
  --mode replace \
  output/titsa_schedules.parquet
```

### 3. Process and Upload (All-in-One)

Process and upload in a single command:

```bash
export TINYBIRD_TOKEN="your_token_here"

./gtfs-processor all \
  --datasource titsa_schedules \
  --mode replace
```

### Options

**Process Command:**
- `--stop-id, -s`: Stop ID to filter (optional, 0 = all stops)
- `--start-date, -b`: Start date YYYYMMDD (optional, 0 = no filter)
- `--end-date, -e`: End date YYYYMMDD (optional, 0 = no filter)
- `--input, -i`: Input directory (default: `../batch_job/Google_transit`)
- `--output, -o`: Output file (default: `output/titsa_schedules.parquet`)
- `--profile, -p`: Enable profiling (memory, CPU, and performance stats)

**Upload Command:**
- `--token, -t`: Tinybird API token (or use `TINYBIRD_TOKEN` env var)
- `--datasource, -d`: Datasource name (default: `titsa_schedules`)
- `--mode, -m`: Upload mode: `create`, `append`, or `replace` (default: `replace`)

### Help

```bash
./gtfs-processor --help
./gtfs-processor process --help
./gtfs-processor upload --help
./gtfs-processor all --help
```

## Input Files

The processor expects the following GTFS CSV files in the input directory:

- `routes.txt` - Bus routes information
- `trips.txt` - Trip schedules
- `calendar_dates.txt` - Service calendar
- `stop_times.txt` - Stop arrival/departure times

## Output Format

The processor generates a flat Parquet file with the following schema:

| Column            | Type       | Description                                      |
|-------------------|------------|--------------------------------------------------|
| line              | string     | Route short name (e.g., "10")                    |
| stop_id           | int32      | Stop identifier                                  |
| estimated_arrival | timestamp  | Timestamp in GMT (stored as milliseconds)        |

The `estimated_arrival` timestamp combines the calendar date and arrival time into a proper Parquet timestamp type in GMT timezone (Canary Islands standard time). This ensures Tinybird correctly interprets it as datetime64.

## Time Normalization

The processor handles GTFS times that extend beyond 24:00:00 (used for late-night services that technically arrive the next day):

- `25:00:00` → Next day at `01:00:00`
- `26:30:00` → Next day at `02:30:00`
- `29:59:00` → Next day at `05:59:00`

The Unix timestamp is automatically adjusted to reflect the correct date and time, making it easy to query and compare arrival times.

## Upload to Tinybird

The upload functionality is integrated into the CLI. See the "Upload Only" and "Process and Upload" sections above.

You can also use the bash script if preferred:

```bash
chmod +x upload.sh
./upload.sh output/titsa_schedules.parquet titsa_schedules
```

## Performance

- Processes ~880K stop_times records in under 30 seconds
- Memory usage: < 10MB peak (< 200MB for full dataset)
- Output file size: Highly compressed (typically 10-20% of CSV size)

### Profiling

Enable profiling to see detailed performance metrics:

```bash
./gtfs-processor process --profile --stop-id 1918 --start-date 20251103 --end-date 20251103
```

**Metrics Provided:**
- **Duration**: Total execution time
- **Min/Max/Avg Memory**: Memory usage statistics
- **Total Allocated**: Lifetime memory allocation
- **GC Runs**: Garbage collection cycles
- **Goroutines**: Concurrent execution threads
- **CPU Cores**: Available processing cores

**Example Output:**
```
Metric          | Value
Duration        | 160.00 ms
Min Memory      | 414.02 KiB
Max Memory      | 8.78 MiB
Avg Memory      | 8.78 MiB
Final Memory    | 468.97 KiB
Total Allocated | 105.60 MiB
Sys Memory      | 21.27 MiB
GC Runs         | 25
Goroutines      | 16
CPU Cores       | 11
```

## Example

```bash
# Process and upload all data in one command
export TINYBIRD_TOKEN="your_token_here"
./gtfs-processor all

# Or process with filters
./gtfs-processor process --stop-id 1918 --start-date 20221018 --end-date 20221020

# Output:
╔══════════════════════════════════════════╗
║  TITSA GTFS Processor                   ║
╚══════════════════════════════════════════╝

ℹ Filtering by stop ID: 1918
ℹ Date range: 20221018 to 20221020

✔ Loaded 176 routes
✔ Loaded 32329 trips
✔ Loaded 1247 calendar dates
ℹ Processed 100000 records, matched 142...
ℹ Processed 200000 records, matched 287...
...
✔ Processed 432 records

╔══════════════════════════════════════════╗
║              ✓ Success                  ║
╚══════════════════════════════════════════╝
Output file: output/titsa_schedules.parquet
Total records: 432

# Then upload
./gtfs-processor upload
```

## Testing

The project includes comprehensive unit tests covering all major functionality:

```bash
# Run all tests
go test -v

# Run with coverage
go test -v -cover

# Run specific test
go test -v -run TestConvertToTimestamp
```

**Test Coverage:**
- ✅ CLI argument parsing
- ✅ Data loading and validation
- ✅ GTFS time normalization (24:00-29:59)
- ✅ Data relationship integrity
- ✅ Tinybird upload structure

See [TESTING.md](TESTING.md) for detailed test documentation.

## Development

### Project Structure

```
gtfs-processor/
├── main.go           # CLI setup and entry point
├── commands.go       # Command handlers (process, upload, all)
├── models.go         # Data structures
├── loader.go         # CSV loading functions
├── processor.go      # GTFS processing logic
├── uploader.go       # Tinybird API integration
├── go.mod            # Go module definition
├── go.sum            # Go dependencies
├── upload.sh         # Tinybird upload script (alternative)
├── README.md         # This file
└── output/           # Generated output files
```

The code is organized into logical modules:
- **main.go**: CLI definition and flags
- **commands.go**: Command execution logic with nice UI
- **models.go**: Type definitions for GTFS entities
- **loader.go**: Functions to read and parse CSV files
- **processor.go**: Core business logic for joining and transforming data
- **uploader.go**: HTTP client for Tinybird API

### Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/pterm/pterm` - Beautiful terminal output
- `github.com/xitongsys/parquet-go` - Parquet file support

## License

MIT

## Author

Adrian - TITSA Ingestion Project

