package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"raznar.id/proxmox-traffic-monitor/internal/storage"
)

var jsonOutput bool
var force bool

var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Fetch all traffic records for a specific VM or for all VMs.",
	Long:  `Fetch all traffic records for a specific VM (if an ID is provided) or for all VMs (if no ID is provided).`,
	Run:   getTraffic,
}

var getDailyCmd = &cobra.Command{
	Use:   "get-daily [id]",
	Short: "Fetch daily traffic for a specific VM or for all VMs.",
	Long:  `Fetch daily traffic for a specific VM (if an ID is provided) or for all VMs (if no ID is provided). The date defaults to today.`,
	Run:   getDailyTraffic,
}

var getMonthlyCmd = &cobra.Command{
	Use:   "get-monthly [id]",
	Short: "Fetch monthly traffic for a specific VM or for all VMs.",
	Long:  `Fetch monthly traffic for a specific VM (if an ID is provided) or for all VMs (if no ID is provided). The month defaults to the current month.`,
	Run:   getMonthlyTraffic,
}

var clearCmd = &cobra.Command{
	Use:   "clear [id]",
	Short: "Clear traffic data for a specific VM or for all VMs.",
	Long:  `Clear traffic data for a specific VM (if an ID is provided) or for all VMs (if no ID is provided).`,
	Run:   clearTraffic,
}

func getTraffic(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	records, err := db.GetTraffic(id)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get traffic data")
	}

	printRecords(records, true)
}

func getDailyTraffic(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	date, _ := cmd.Flags().GetString("date")
	if date == "" {
		date = time.Now().Format("02-01-06")
	}
	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	records, err := db.GetDailyTraffic(id, date)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get daily traffic data")
	}

	printRecords(records, false)
}

func getMonthlyTraffic(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	month, _ := cmd.Flags().GetString("month")
	if month == "" {
		month = time.Now().Format("-01-06")
	}
	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	records, err := db.GetMonthlyTraffic(id, month)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get monthly traffic data")
	}

	printRecords(records, false)
}

func clearTraffic(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	if !force {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure you want to clear the traffic data? (y/n): ")
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) != "y" {
			fmt.Println("Operation cancelled.")
			return
		}
	}
	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	err = db.ClearTraffic(id)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to clear traffic data")
	}

	fmt.Println("Traffic data cleared successfully.")
}

func printRecords(records []storage.TrafficRecord, hideDate bool) {
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(records)
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)

	// Header
	if hideDate {
		fmt.Fprintln(w, "ID\tIN (MB)\tOUT (MB)")
	} else {
		fmt.Fprintln(w, "ID\tDATE\tIN (MB)\tOUT (MB)")
	}

	for _, r := range records {
		inMB := float64(r.In) / 1024 / 1024
		outMB := float64(r.Out) / 1024 / 1024

		if hideDate {
			fmt.Fprintf(w, "%s\t%.2f\t%.2f\n", r.ID, inMB, outMB)
		} else {
			// Parse stored date and format it
			displayDate := r.Date
			if parsed, err := time.Parse("02-01-06", r.Date); err == nil {
				displayDate = parsed.Format("02 Jan 2006") // human-friendly
			}
			fmt.Fprintf(w, "%s\t%s\t%.2f\t%.2f\n", r.ID, displayDate, inMB, outMB)
		}
	}

	w.Flush()
}

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(getDailyCmd)
	rootCmd.AddCommand(getMonthlyCmd)
	rootCmd.AddCommand(clearCmd)

	getCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getDailyCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getDailyCmd.Flags().String("date", "", "Date in DD-MM-YY format (defaults to today)")
	getMonthlyCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getMonthlyCmd.Flags().String("month", "", "Month in -MM-YY format (defaults to current month)")
	clearCmd.Flags().BoolVar(&force, "force", false, "Force clear without confirmation")
}
