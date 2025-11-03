package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// loadRoutes reads the routes.txt file and returns a map of route_id -> Route
func loadRoutes(filename string) (map[string]*Route, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	header = cleanHeader(header)

	routes := make(map[string]*Route)

	routeIDIdx := findColumnIndex(header, "route_id")
	routeShortNameIdx := findColumnIndex(header, "route_short_name")

	if routeIDIdx == -1 || routeShortNameIdx == -1 {
		return nil, fmt.Errorf("missing required columns in routes.txt (found: %v)", header)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		route := &Route{
			RouteID:        record[routeIDIdx],
			RouteShortName: record[routeShortNameIdx],
		}
		routes[route.RouteID] = route
	}

	return routes, nil
}

// loadTrips reads the trips.txt file and returns a map of trip_id -> Trip
func loadTrips(filename string) (map[string]*Trip, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	header = cleanHeader(header)

	trips := make(map[string]*Trip)

	routeIDIdx := findColumnIndex(header, "route_id")
	serviceIDIdx := findColumnIndex(header, "service_id")
	tripIDIdx := findColumnIndex(header, "trip_id")

	if routeIDIdx == -1 || serviceIDIdx == -1 || tripIDIdx == -1 {
		return nil, fmt.Errorf("missing required columns in trips.txt (found: %v)", header)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		trip := &Trip{
			RouteID:   record[routeIDIdx],
			ServiceID: record[serviceIDIdx],
			TripID:    record[tripIDIdx],
		}
		trips[trip.TripID] = trip
	}

	return trips, nil
}

// loadCalendarDates reads the calendar_dates.txt file and returns a map of service_id -> []dates
// Filters dates based on the global startDate and endDate variables
func loadCalendarDates(filename string, startDate, endDate int) (map[string][]int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	header = cleanHeader(header)

	calendarDates := make(map[string][]int)

	serviceIDIdx := findColumnIndex(header, "service_id")
	dateIdx := findColumnIndex(header, "date")

	if serviceIDIdx == -1 || dateIdx == -1 {
		return nil, fmt.Errorf("missing required columns in calendar_dates.txt (found: %v)", header)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		serviceID := record[serviceIDIdx]
		date, err := strconv.Atoi(record[dateIdx])
		if err != nil {
			continue
		}

		// Filter dates in range (if filters are set)
		includeDate := true
		if startDate > 0 && date < startDate {
			includeDate = false
		}
		if endDate > 0 && date > endDate {
			includeDate = false
		}

		if includeDate {
			calendarDates[serviceID] = append(calendarDates[serviceID], date)
		}
	}

	return calendarDates, nil
}

// cleanHeader removes BOM and trims whitespace from CSV headers
func cleanHeader(header []string) []string {
	cleaned := make([]string, len(header))
	for i, col := range header {
		cleaned[i] = strings.TrimSpace(col)
		cleaned[i] = strings.TrimPrefix(cleaned[i], "\ufeff") // Remove UTF-8 BOM
	}
	return cleaned
}

// findColumnIndex returns the index of a column name in the header, or -1 if not found
func findColumnIndex(header []string, columnName string) int {
	for i, col := range header {
		if col == columnName {
			return i
		}
	}
	return -1
}
