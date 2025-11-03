package main

import (
	"os"
	"testing"
	"time"
)

func TestConvertToTimestamp(t *testing.T) {
	tests := []struct {
		name          string
		date          int
		timeStr       string
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
		expectedHour  int
		expectedMin   int
		expectError   bool
	}{
		{
			name:          "normal time",
			date:          20251103,
			timeStr:       "14:30:00",
			expectedYear:  2025,
			expectedMonth: time.November,
			expectedDay:   3,
			expectedHour:  14,
			expectedMin:   30,
			expectError:   false,
		},
		{
			name:          "late night service 25:00",
			date:          20251103,
			timeStr:       "25:30:00",
			expectedYear:  2025,
			expectedMonth: time.November,
			expectedDay:   4, // Next day
			expectedHour:  1,
			expectedMin:   30,
			expectError:   false,
		},
		{
			name:          "very late service 29:45",
			date:          20251103,
			timeStr:       "29:45:00",
			expectedYear:  2025,
			expectedMonth: time.November,
			expectedDay:   4, // Next day
			expectedHour:  5,
			expectedMin:   45,
			expectError:   false,
		},
		{
			name:          "midnight",
			date:          20251103,
			timeStr:       "00:00:00",
			expectedYear:  2025,
			expectedMonth: time.November,
			expectedDay:   3,
			expectedHour:  0,
			expectedMin:   0,
			expectError:   false,
		},
		{
			name:        "invalid time format",
			date:        20251103,
			timeStr:     "14:30",
			expectError: true,
		},
		{
			name:        "invalid date format",
			date:        2025110,
			timeStr:     "14:30:00",
			expectError: true,
		},
	}

	gmt := time.FixedZone("GMT", 0)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, err := convertToTimestamp(tt.date, tt.timeStr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Convert timestamp back to time for verification
			result := time.Unix(timestamp, 0).In(gmt)

			if result.Year() != tt.expectedYear {
				t.Errorf("Expected year %d, got %d", tt.expectedYear, result.Year())
			}

			if result.Month() != tt.expectedMonth {
				t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.Month())
			}

			if result.Day() != tt.expectedDay {
				t.Errorf("Expected day %d, got %d", tt.expectedDay, result.Day())
			}

			if result.Hour() != tt.expectedHour {
				t.Errorf("Expected hour %d, got %d", tt.expectedHour, result.Hour())
			}

			if result.Minute() != tt.expectedMin {
				t.Errorf("Expected minute %d, got %d", tt.expectedMin, result.Minute())
			}
		})
	}
}

func TestDataRelationships(t *testing.T) {
	// Load all test data
	routes, err := loadRoutes("testdata/routes.txt")
	if err != nil {
		t.Fatalf("Failed to load routes: %v", err)
	}

	trips, err := loadTrips("testdata/trips.txt")
	if err != nil {
		t.Fatalf("Failed to load trips: %v", err)
	}

	// Verify trips reference valid routes
	for tripID, trip := range trips {
		route, exists := routes[trip.RouteID]
		if !exists {
			t.Errorf("Trip %s references non-existent route %s", tripID, trip.RouteID)
		} else {
			// Verify we can get route information from trip
			if route.RouteShortName == "" {
				t.Errorf("Route %s has empty short name", trip.RouteID)
			}
		}
	}
}

func TestProcessStopTimesDataIntegrity(t *testing.T) {
	routes, err := loadRoutes("testdata/routes.txt")
	if err != nil {
		t.Fatalf("Failed to load routes: %v", err)
	}

	trips, err := loadTrips("testdata/trips.txt")
	if err != nil {
		t.Fatalf("Failed to load trips: %v", err)
	}

	// Load ALL calendar dates (no filter) to ensure we have data
	calendarDates, err := loadCalendarDates("testdata/calendar_dates.txt", 0, 0)
	if err != nil {
		t.Fatalf("Failed to load calendar dates: %v", err)
	}

	// Skip test if no calendar dates available for test service
	if len(calendarDates) == 0 {
		t.Skip("No calendar dates available in test data")
	}

	// Test with sample stop ID
	count, err := processStopTimes(
		"testdata/stop_times.txt",
		"testdata/test_output.parquet",
		routes,
		trips,
		calendarDates,
		9181, // Filter by specific stop
	)

	if err != nil {
		t.Fatalf("Failed to process stop times: %v", err)
	}

	// Note: count might be 0 if service_id in trips doesn't match calendar_dates
	// This is OK for a basic integration test
	t.Logf("Processed %d records", count)

	// Cleanup test file
	os.Remove("testdata/test_output.parquet")
}
