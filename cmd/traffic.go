package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
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

var getRangeCmd = &cobra.Command{
	Use:   "get-range [id]",
	Short: "Fetch traffic records for a specific VM or all VMs within a date range",
	Long:  `Fetch traffic records for a specific VM (if an ID is provided) or for all VMs (if no ID is provided) between start and end dates.`,
	Run:   getRangeTraffic,
}

var getRangeTotalCmd = &cobra.Command{
	Use:   "get-range-total [id]",
	Short: "Fetch total traffic for a VM or all VMs within a date range",
	Long:  `Fetch the sum of all traffic (IN/OUT) for a specific VM (if ID is provided) or all VMs (if no ID) between start and end dates.`,
	Run:   getRangeTotalTraffic,
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

	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	exists, err := db.ExistsTraffic(id)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check traffic existence")
	}
	if !exists {
		fmt.Println("No traffic data found for the specified VM.")
		return
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

	err = db.ClearTraffic(id)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to clear traffic data")
	}

	fmt.Println("Traffic data cleared successfully.")
}

func getRangeTraffic(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	startDate, _ := cmd.Flags().GetString("start")
	endDate, _ := cmd.Flags().GetString("end")

	if startDate == "" || endDate == "" {
		log.Fatal().Msg("Both --start and --end dates must be provided in DD-MM-YYYY format")
	}

	// Parse dates with 4-digit year
	const layout = "02-01-2006"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		log.Fatal().Err(err).Msgf("Invalid --start date format, expected DD-MM-YYYY, got %s", startDate)
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		log.Fatal().Err(err).Msgf("Invalid --end date format, expected DD-MM-YYYY, got %s", endDate)
	}

	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	records, err := db.GetTrafficByRange(id, start, end)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get traffic data by range")
	}

	printRecords(records, false)
}

func getRangeTotalTraffic(cmd *cobra.Command, args []string) {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}

	startDate, _ := cmd.Flags().GetString("start")
	endDate, _ := cmd.Flags().GetString("end")

	if startDate == "" || endDate == "" {
		log.Fatal().Msg("Both --start and --end dates must be provided in DD-MM-YYYY format")
	}

	const layout = "02-01-2006"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		log.Fatal().Err(err).Msgf("Invalid --start date format, expected DD-MM-YYYY, got %s", startDate)
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		log.Fatal().Err(err).Msgf("Invalid --end date format, expected DD-MM-YYYY, got %s", endDate)
	}

	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	records, err := db.GetTrafficTotalByRange(id, start, end)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get total traffic by range")
	}

	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(records)
		return
	}

	printRecords(records, true)
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
		fmt.Fprintln(w, "ID\tVMID\tNODEID\tIN\tOUT")
	} else {
		fmt.Fprintln(w, "ID\tVMID\tNODEID\tDATE\tIN\tOUT")
	}

	for _, r := range records {
		inHuman := humanize.Bytes(r.In) // automatically converts bytes to KB, MB, GB
		outHuman := humanize.Bytes(r.Out)

		if hideDate {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n", r.ID, r.VMID, r.NodeID, inHuman, outHuman)
		} else {
			displayDate := r.Date
			if parsed, err := time.Parse("02-01-06", r.Date); err == nil {
				displayDate = parsed.Format("02 Jan 2006") // human-friendly
			}
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\n", r.ID, r.VMID, r.NodeID, displayDate, inHuman, outHuman)
		}
	}

	w.Flush()
}


func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(getDailyCmd)
	rootCmd.AddCommand(getMonthlyCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(getRangeCmd)
	rootCmd.AddCommand(getRangeTotalCmd)

	getRangeTotalCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getRangeTotalCmd.Flags().String("start", "", "Start date in DD-MM-YYYY format")
	getRangeTotalCmd.Flags().String("end", "", "End date in DD-MM-YYYY format")
	getRangeCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getRangeCmd.Flags().String("start", "", "Start date in DD-MM-YY format")
	getRangeCmd.Flags().String("end", "", "End date in DD-MM-YY format")
	getCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getDailyCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getDailyCmd.Flags().String("date", "", "Date in DD-MM-YY format (defaults to today)")
	getMonthlyCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	getMonthlyCmd.Flags().String("month", "", "Month in -MM-YY format (defaults to current month)")
	clearCmd.Flags().BoolVar(&force, "force", false, "Force clear without confirmation")
}
