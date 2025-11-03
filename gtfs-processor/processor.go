package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"
)

// processStopTimes reads stop_times.txt, joins with other data, and writes to Parquet
func processStopTimes(
	filename string,
	outputFile string,
	routes map[string]*Route,
	trips map[string]*Trip,
	calendarDates map[string][]int,
	stopIDFilter int,
) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Create Parquet writer
	fw, err := local.NewLocalFileWriter(outputFile)
	if err != nil {
		return 0, err
	}
	defer fw.Close()

	pw, err := writer.NewParquetWriter(fw, new(TitsaArrival), 4)
	if err != nil {
		return 0, err
	}
	defer pw.WriteStop()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return 0, err
	}

	header = cleanHeader(header)

	tripIDIdx := findColumnIndex(header, "trip_id")
	arrivalTimeIdx := findColumnIndex(header, "arrival_time")
	stopIDIdx := findColumnIndex(header, "stop_id")

	if tripIDIdx == -1 || arrivalTimeIdx == -1 || stopIDIdx == -1 {
		return 0, fmt.Errorf("missing required columns in stop_times.txt (found: %v)", header)
	}

	count := 0
	recordCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}

		recordCount++
		if recordCount%10000 == 0 {
			pterm.Info.Printf("Processed %d records, matched %d...\n", recordCount, count)
		}

		// Parse stop_id
		stopIDStr := record[stopIDIdx]
		currentStopID, err := strconv.Atoi(stopIDStr)
		if err != nil {
			continue
		}

		// Filter by stop_id (if filter is set)
		if stopIDFilter > 0 && currentStopID != stopIDFilter {
			continue
		}

		tripID := record[tripIDIdx]
		trip, exists := trips[tripID]
		if !exists {
			continue
		}

		// Check if service has dates in range
		dates, exists := calendarDates[trip.ServiceID]
		if !exists || len(dates) == 0 {
			continue
		}

		// Get route info
		route, exists := routes[trip.RouteID]
		if !exists {
			continue
		}

		arrivalTime := record[arrivalTimeIdx]

		// Process each date for this service
		for _, date := range dates {
			timestamp, err := convertToTimestamp(date, arrivalTime)
			if err != nil {
				// Skip invalid timestamps
				continue
			}

			// Convert to milliseconds for Parquet TIMESTAMP_MILLIS
			timestampMillis := timestamp * 1000

			arrival := &TitsaArrival{
				Line:             route.RouteShortName,
				StopID:           int32(currentStopID),
				EstimatedArrival: timestampMillis,
			}

			if err := pw.Write(arrival); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}

// convertToTimestamp converts GTFS date (YYYYMMDD) and time (HH:MM:SS) to Unix timestamp
// Handles GTFS times like 25:00:00 (next day 01:00:00)
// Uses GMT timezone (Canary Islands)
func convertToTimestamp(date int, timeStr string) (int64, error) {
	// Parse time string
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	second, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid second: %s", parts[2])
	}

	// Parse date (YYYYMMDD format)
	dateStr := strconv.Itoa(date)
	if len(dateStr) != 8 {
		return 0, fmt.Errorf("invalid date format: %d", date)
	}

	year, err := strconv.Atoi(dateStr[0:4])
	if err != nil {
		return 0, fmt.Errorf("invalid year: %s", dateStr[0:4])
	}

	month, err := strconv.Atoi(dateStr[4:6])
	if err != nil {
		return 0, fmt.Errorf("invalid month: %s", dateStr[4:6])
	}

	day, err := strconv.Atoi(dateStr[6:8])
	if err != nil {
		return 0, fmt.Errorf("invalid day: %s", dateStr[6:8])
	}

	// Handle GTFS times >= 24:00:00 (next day service)
	daysToAdd := 0
	if hour >= 24 {
		daysToAdd = hour / 24
		hour = hour % 24
	}

	// Create timestamp in GMT timezone (Canary Islands)
	gmt := time.FixedZone("GMT", 0)
	timestamp := time.Date(year, time.Month(month), day, hour, minute, second, 0, gmt)

	// Add days if needed for late-night services
	if daysToAdd > 0 {
		timestamp = timestamp.AddDate(0, 0, daysToAdd)
	}

	return timestamp.Unix(), nil
}
