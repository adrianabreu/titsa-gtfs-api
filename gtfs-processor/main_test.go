package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCLIFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		checkStopID   bool
		expectedStop  int
		checkDates    bool
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "process with all flags",
			args:          []string{"process", "--stop-id", "1918", "--start-date", "20251103", "--end-date", "20251105"},
			expectError:   false,
			checkStopID:   true,
			expectedStop:  1918,
			checkDates:    true,
			expectedStart: 20251103,
			expectedEnd:   20251105,
		},
		{
			name:         "process without flags (defaults)",
			args:         []string{"process"},
			expectError:  false,
			checkStopID:  true,
			expectedStop: 0, // Default
		},
		{
			name:         "process with stop ID only",
			args:         []string{"process", "--stop-id", "1918"},
			expectError:  false,
			checkStopID:  true,
			expectedStop: 1918,
		},
		{
			name:          "process with dates only",
			args:          []string{"process", "--start-date", "20251103", "--end-date", "20251105"},
			expectError:   false,
			checkDates:    true,
			expectedStart: 20251103,
			expectedEnd:   20251105,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			stopID = 0
			startDate = 0
			endDate = 0
			inputDir = ""
			outputFile = ""

			// Create a new root command for testing
			rootCmd := &cobra.Command{Use: "gtfs-processor"}

			processCmd := &cobra.Command{
				Use: "process",
				Run: func(cmd *cobra.Command, args []string) {
					// Don't actually run, just parse flags
				},
			}

			processCmd.Flags().IntVarP(&stopID, "stop-id", "s", 0, "Stop ID to filter")
			processCmd.Flags().IntVarP(&startDate, "start-date", "b", 0, "Start date")
			processCmd.Flags().IntVarP(&endDate, "end-date", "e", 0, "End date")
			processCmd.Flags().StringVarP(&inputDir, "input", "i", "../batch_job/Google_transit", "Input directory")
			processCmd.Flags().StringVarP(&outputFile, "output", "o", "output/titsa_schedules.parquet", "Output file")

			rootCmd.AddCommand(processCmd)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.checkStopID && stopID != tt.expectedStop {
				t.Errorf("Expected stop ID %d, got %d", tt.expectedStop, stopID)
			}

			if tt.checkDates {
				if startDate != tt.expectedStart {
					t.Errorf("Expected start date %d, got %d", tt.expectedStart, startDate)
				}
				if endDate != tt.expectedEnd {
					t.Errorf("Expected end date %d, got %d", tt.expectedEnd, endDate)
				}
			}
		})
	}
}

func TestCLIUploadFlags(t *testing.T) {
	// Reset flags
	tinybirdToken = ""
	tinybirdDatasource = ""
	uploadMode = ""

	rootCmd := &cobra.Command{Use: "gtfs-processor"}

	uploadCmd := &cobra.Command{
		Use: "upload",
		Run: func(cmd *cobra.Command, args []string) {
			// Don't actually run
		},
	}

	uploadCmd.Flags().StringVarP(&tinybirdToken, "token", "t", "", "Tinybird API token")
	uploadCmd.Flags().StringVarP(&tinybirdDatasource, "datasource", "d", "titsa_schedules", "Datasource name")
	uploadCmd.Flags().StringVarP(&uploadMode, "mode", "m", "replace", "Upload mode")

	rootCmd.AddCommand(uploadCmd)
	rootCmd.SetArgs([]string{"upload", "--token", "test_token", "--datasource", "my_data", "--mode", "append"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if tinybirdToken != "test_token" {
		t.Errorf("Expected token 'test_token', got '%s'", tinybirdToken)
	}

	if tinybirdDatasource != "my_data" {
		t.Errorf("Expected datasource 'my_data', got '%s'", tinybirdDatasource)
	}

	if uploadMode != "append" {
		t.Errorf("Expected mode 'append', got '%s'", uploadMode)
	}
}

func TestGetToken(t *testing.T) {
	tests := []struct {
		name       string
		flagToken  string
		envToken   string
		expectExit bool
	}{
		{
			name:       "token from flag",
			flagToken:  "flag_token",
			envToken:   "",
			expectExit: false,
		},
		{
			name:       "token from env",
			flagToken:  "",
			envToken:   "env_token",
			expectExit: false,
		},
		{
			name:       "flag takes precedence",
			flagToken:  "flag_token",
			envToken:   "env_token",
			expectExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envToken != "" {
				os.Setenv("TINYBIRD_TOKEN", tt.envToken)
				defer os.Unsetenv("TINYBIRD_TOKEN")
			} else {
				os.Unsetenv("TINYBIRD_TOKEN")
			}

			tinybirdToken = tt.flagToken

			token := getToken()

			expectedToken := tt.flagToken
			if expectedToken == "" {
				expectedToken = tt.envToken
			}

			if token != expectedToken {
				t.Errorf("Expected token '%s', got '%s'", expectedToken, token)
			}
		})
	}
}

func TestCLIHelp(t *testing.T) {
	rootCmd := &cobra.Command{Use: "gtfs-processor"}
	processCmd := &cobra.Command{Use: "process"}
	rootCmd.AddCommand(processCmd)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "gtfs-processor") {
		t.Error("Help output should contain command name")
	}
}
