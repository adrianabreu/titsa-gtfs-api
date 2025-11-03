package main

import (
	"testing"
)

func TestLoadRoutes(t *testing.T) {
	routes, err := loadRoutes("testdata/routes.txt")
	if err != nil {
		t.Fatalf("Failed to load routes: %v", err)
	}

	if len(routes) == 0 {
		t.Fatal("Expected routes to be loaded, got 0")
	}

	// Test specific route exists
	route, exists := routes["247"]
	if !exists {
		t.Error("Expected route 247 to exist")
	}

	if route.RouteShortName != "10" {
		t.Errorf("Expected route short name '10', got '%s'", route.RouteShortName)
	}

	// Test another route
	route2, exists := routes["13"]
	if !exists {
		t.Error("Expected route 13 to exist")
	}

	if route2.RouteShortName != "11" {
		t.Errorf("Expected route short name '11', got '%s'", route2.RouteShortName)
	}
}

func TestLoadTrips(t *testing.T) {
	trips, err := loadTrips("testdata/trips.txt")
	if err != nil {
		t.Fatalf("Failed to load trips: %v", err)
	}

	if len(trips) == 0 {
		t.Fatal("Expected trips to be loaded, got 0")
	}

	// Test specific trip
	trip, exists := trips["10108802"]
	if !exists {
		t.Error("Expected trip 10108802 to exist")
	}

	if trip.RouteID != "247" {
		t.Errorf("Expected route_id '247', got '%s'", trip.RouteID)
	}

	if trip.ServiceID != "298786" {
		t.Errorf("Expected service_id '298786', got '%s'", trip.ServiceID)
	}
}

func TestLoadCalendarDates(t *testing.T) {
	tests := []struct {
		name        string
		startDate   int
		endDate     int
		minExpected int
	}{
		{
			name:        "no filter",
			startDate:   0,
			endDate:     0,
			minExpected: 10, // All test dates
		},
		{
			name:        "with filter",
			startDate:   20251103,
			endDate:     20251103,
			minExpected: 0, // May or may not have dates in this range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calDates, err := loadCalendarDates("testdata/calendar_dates.txt", tt.startDate, tt.endDate)
			if err != nil {
				t.Fatalf("Failed to load calendar dates: %v", err)
			}

			totalDates := 0
			for _, dates := range calDates {
				totalDates += len(dates)
			}

			if totalDates < tt.minExpected {
				t.Errorf("Expected at least %d dates, got %d", tt.minExpected, totalDates)
			}
		})
	}
}

func TestCleanHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "normal headers",
			input:    []string{"route_id", "agency_id", "route_short_name"},
			expected: []string{"route_id", "agency_id", "route_short_name"},
		},
		{
			name:     "headers with spaces",
			input:    []string{"  route_id  ", "agency_id", "route_short_name  "},
			expected: []string{"route_id", "agency_id", "route_short_name"},
		},
		{
			name:     "headers with BOM",
			input:    []string{"\ufeffroute_id", "agency_id"},
			expected: []string{"route_id", "agency_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanHeader(tt.input)
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected '%s', got '%s'", expected, result[i])
				}
			}
		})
	}
}

func TestFindColumnIndex(t *testing.T) {
	header := []string{"route_id", "agency_id", "route_short_name"}

	tests := []struct {
		name       string
		columnName string
		expected   int
	}{
		{"first column", "route_id", 0},
		{"middle column", "agency_id", 1},
		{"last column", "route_short_name", 2},
		{"non-existent", "invalid_column", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findColumnIndex(header, tt.columnName)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
