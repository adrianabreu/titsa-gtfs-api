package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// runProcessor handles the "process" command
func runProcessor(cmd *cobra.Command, args []string) {
	// Start profiler if enabled
	profiler := NewProfiler()
	profiler.Start()
	defer profiler.Stop()

	printBanner("TITSA GTFS Processor")

	// Display filter info
	displayFilters()

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		pterm.Error.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Load data files
	routes, trips, calendarDates := loadAllData()

	// Process stop times and write output
	spinnerInfo, _ := pterm.DefaultSpinner.Start("Processing stop times and generating Parquet file...")
	count, err := processStopTimes(
		filepath.Join(inputDir, "stop_times.txt"),
		outputFile,
		routes,
		trips,
		calendarDates,
		stopID,
	)
	if err != nil {
		spinnerInfo.Fail(fmt.Sprintf("Failed to process stop times: %v", err))
		os.Exit(1)
	}
	spinnerInfo.Success(fmt.Sprintf("Processed %d records", count))

	// Success message
	pterm.Println()
	displaySuccessBox("Success", fmt.Sprintf("Output file: %s\nTotal records: %d", outputFile, count))
}

// runUpload handles the "upload" command
func runUpload(cmd *cobra.Command, args []string) {
	printBanner("Tinybird Upload")

	// Determine file to upload
	fileToUpload := "output/titsa_schedules.parquet"
	if len(args) > 0 {
		fileToUpload = args[0]
	}

	// Get token
	token := getToken()

	// Check if file exists
	fileInfo, err := os.Stat(fileToUpload)
	if err != nil {
		pterm.Error.Printf("File not found: %s\n", fileToUpload)
		os.Exit(1)
	}

	// Display upload info
	displayUploadInfo(fileToUpload, fileInfo.Size())

	// Upload
	spinnerInfo, _ := pterm.DefaultSpinner.Start("Uploading to Tinybird...")

	err = uploadToTinybird(fileToUpload, token, tinybirdDatasource, uploadMode)
	if err != nil {
		spinnerInfo.Fail(fmt.Sprintf("Upload failed: %v", err))
		os.Exit(1)
	}

	spinnerInfo.Success("Upload completed successfully!")

	// Success message
	pterm.Println()
	displaySuccessBox("Success", fmt.Sprintf("Data uploaded to Tinybird datasource: %s", tinybirdDatasource))
}

// runAll handles the "all" command (process + upload)
func runAll(cmd *cobra.Command, args []string) {
	// Get token first
	token := getToken()

	// Run processor
	runProcessor(cmd, args)

	pterm.Println()
	pterm.Info.Println("Starting upload to Tinybird...")
	pterm.Println()

	// Check output file
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		pterm.Error.Printf("Failed to read output file: %v\n", err)
		os.Exit(1)
	}

	// Display upload info
	displayUploadInfo(outputFile, fileInfo.Size())

	// Upload
	spinnerInfo, _ := pterm.DefaultSpinner.Start("Uploading to Tinybird...")

	err = uploadToTinybird(outputFile, token, tinybirdDatasource, uploadMode)
	if err != nil {
		spinnerInfo.Fail(fmt.Sprintf("Upload failed: %v", err))
		os.Exit(1)
	}

	spinnerInfo.Success("Upload completed successfully!")

	pterm.Println()
	displaySuccessBox("All Done", "Data processed and uploaded to Tinybird successfully!")
}

// loadAllData loads all required GTFS data files
func loadAllData() (map[string]*Route, map[string]*Trip, map[string][]int) {
	var spinnerInfo *pterm.SpinnerPrinter
	var err error

	// Load routes
	spinnerInfo, _ = pterm.DefaultSpinner.Start("Loading routes...")
	routes, err := loadRoutes(filepath.Join(inputDir, "routes.txt"))
	if err != nil {
		spinnerInfo.Fail(fmt.Sprintf("Failed to load routes: %v", err))
		os.Exit(1)
	}
	spinnerInfo.Success(fmt.Sprintf("Loaded %d routes", len(routes)))

	// Load trips
	spinnerInfo, _ = pterm.DefaultSpinner.Start("Loading trips...")
	trips, err := loadTrips(filepath.Join(inputDir, "trips.txt"))
	if err != nil {
		spinnerInfo.Fail(fmt.Sprintf("Failed to load trips: %v", err))
		os.Exit(1)
	}
	spinnerInfo.Success(fmt.Sprintf("Loaded %d trips", len(trips)))

	// Load calendar dates
	spinnerInfo, _ = pterm.DefaultSpinner.Start("Loading calendar dates...")
	calendarDates, err := loadCalendarDates(filepath.Join(inputDir, "calendar_dates.txt"), startDate, endDate)
	if err != nil {
		spinnerInfo.Fail(fmt.Sprintf("Failed to load calendar dates: %v", err))
		os.Exit(1)
	}
	spinnerInfo.Success(fmt.Sprintf("Loaded %d calendar dates", len(calendarDates)))

	return routes, trips, calendarDates
}

// Helper functions for display

func printBanner(title string) {
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println(title)
}

func displayFilters() {
	if stopID > 0 {
		pterm.Info.Printf("Filtering by stop ID: %d\n", stopID)
	} else {
		pterm.Info.Println("Processing ALL stops")
	}

	if startDate > 0 && endDate > 0 {
		pterm.Info.Printf("Date range: %d to %d\n", startDate, endDate)
	} else if startDate > 0 {
		pterm.Info.Printf("Filtering from date: %d\n", startDate)
	} else if endDate > 0 {
		pterm.Info.Printf("Filtering to date: %d\n", endDate)
	} else {
		pterm.Info.Println("Processing ALL dates")
	}

	pterm.Println()
}

func displayUploadInfo(filename string, size int64) {
	pterm.Info.Printf("File: %s\n", filename)
	pterm.Info.Printf("Size: %.2f MB\n", float64(size)/(1024*1024))
	pterm.Info.Printf("Datasource: %s\n", tinybirdDatasource)
	pterm.Info.Printf("Mode: %s\n", uploadMode)
	pterm.Println()
}

func displaySuccessBox(title, message string) {
	pterm.DefaultBox.WithTitle("✓ " + title).
		WithTitleTopCenter().
		WithRightPadding(4).
		WithLeftPadding(4).
		WithBoxStyle(pterm.NewStyle(pterm.FgGreen)).
		Println(message)
}

func getToken() string {
	token := tinybirdToken
	if token == "" {
		token = os.Getenv("TINYBIRD_TOKEN")
	}
	if token == "" {
		pterm.Error.Println("Tinybird token required. Use --token flag or set TINYBIRD_TOKEN environment variable")
		os.Exit(1)
	}
	return token
}
