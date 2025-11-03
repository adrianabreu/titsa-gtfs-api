package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// CLI flags
var (
	// Process flags
	stopID     int
	startDate  int
	endDate    int
	inputDir   string
	outputFile string

	// Upload flags
	tinybirdToken      string
	tinybirdDatasource string
	uploadMode         string

	// Profile flag
	profileEnabled bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gtfs-processor",
		Short: "GTFS data processor for TITSA bus schedules",
		Long: `
╔══════════════════════════════════════════╗
║  TITSA GTFS Data Processor              ║
║  Merge and transform GTFS data          ║
╚══════════════════════════════════════════╝

Processes GTFS CSV files and generates a flat Parquet output
for efficient querying and storage.`,
	}

	// Process command
	processCmd := &cobra.Command{
		Use:   "process",
		Short: "Process GTFS CSV files and generate Parquet output",
		Long:  "Read GTFS CSV files, join and transform data, and output a flat Parquet file",
		Run:   runProcessor,
	}

	processCmd.Flags().IntVarP(&stopID, "stop-id", "s", 0, "Stop ID to filter (optional, 0 = all stops)")
	processCmd.Flags().IntVarP(&startDate, "start-date", "b", 0, "Start date in YYYYMMDD format (optional, 0 = no filter)")
	processCmd.Flags().IntVarP(&endDate, "end-date", "e", 0, "End date in YYYYMMDD format (optional, 0 = no filter)")
	processCmd.Flags().StringVarP(&inputDir, "input", "i", "../batch_job/Google_transit", "Input directory with GTFS CSV files")
	processCmd.Flags().StringVarP(&outputFile, "output", "o", "output/titsa_schedules.parquet", "Output Parquet file path")
	processCmd.Flags().BoolVarP(&profileEnabled, "profile", "p", false, "Enable profiling (memory and CPU statistics)")

	// Upload command
	uploadCmd := &cobra.Command{
		Use:   "upload [parquet-file]",
		Short: "Upload Parquet file to Tinybird",
		Long:  "Upload the generated Parquet file to Tinybird datasource via API",
		Args:  cobra.MaximumNArgs(1),
		Run:   runUpload,
	}

	uploadCmd.Flags().StringVarP(&tinybirdToken, "token", "t", "", "Tinybird API token (or set TINYBIRD_TOKEN env var)")
	uploadCmd.Flags().StringVarP(&tinybirdDatasource, "datasource", "d", "titsa_schedules", "Tinybird datasource name")
	uploadCmd.Flags().StringVarP(&uploadMode, "mode", "m", "replace", "Upload mode: create, append, or replace")

	// All-in-one command (process and upload)
	allCmd := &cobra.Command{
		Use:   "all",
		Short: "Process GTFS files and upload to Tinybird in one step",
		Long:  "Execute both process and upload commands sequentially",
		Run:   runAll,
	}

	allCmd.Flags().IntVarP(&stopID, "stop-id", "s", 0, "Stop ID to filter (optional, 0 = all stops)")
	allCmd.Flags().IntVarP(&startDate, "start-date", "b", 0, "Start date in YYYYMMDD format (optional, 0 = no filter)")
	allCmd.Flags().IntVarP(&endDate, "end-date", "e", 0, "End date in YYYYMMDD format (optional, 0 = no filter)")
	allCmd.Flags().StringVarP(&inputDir, "input", "i", "../batch_job/Google_transit", "Input directory with GTFS CSV files")
	allCmd.Flags().StringVarP(&outputFile, "output", "o", "output/titsa_schedules.parquet", "Output Parquet file path")
	allCmd.Flags().StringVarP(&tinybirdToken, "token", "t", "", "Tinybird API token (or set TINYBIRD_TOKEN env var)")
	allCmd.Flags().StringVarP(&tinybirdDatasource, "datasource", "d", "titsa_schedules", "Tinybird datasource name")
	allCmd.Flags().StringVarP(&uploadMode, "mode", "m", "replace", "Upload mode: create, append, or replace")
	allCmd.Flags().BoolVarP(&profileEnabled, "profile", "p", false, "Enable profiling (memory and CPU statistics)")

	rootCmd.AddCommand(processCmd, uploadCmd, allCmd)

	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
