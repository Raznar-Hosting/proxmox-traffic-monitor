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

	dateStr, _ := cmd.Flags().GetString("date")
	var day time.Time
	var err error

	if dateStr == "" {
		day = time.Now()
	} else {
		// Expecting DD-MM-YYYY
		day, err = time.Parse("02-01-2006", dateStr)
		if err != nil {
			log.Fatal().Err(err).Msgf("Invalid date format, expected DD-MM-YYYY, got %s", dateStr)
		}
	}

	appConfig := loadConfig()
	db, err := storage.New(appConfig.App.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	records, err := db.GetDailyTraffic(id, day)
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

	monthStr, _ := cmd.Flags().GetString("month")
	var month string

	if monthStr == "" {
		month = time.Now().Format("2006-01") // YYYY-MM
	} else {
		// Validate format YYYY-MM
		if _, err := time.Parse("2006-01", monthStr); err != nil {
			log.Fatal().Err(err).Msgf("Invalid month format, expected YYYY-MM, got %s", monthStr)
		}
		month = monthStr
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

	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")

	const layout = "02-01-2006"

	var start *time.Time
	var end *time.Time

	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Invalid --start date format, expected DD-MM-YYYY, got %s", startStr)
		}
		start = &t
	}

	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Invalid --end date format, expected DD-MM-YYYY, got %s", endStr)
		}
		// include full day
		t = t.AddDate(0, 0, 1)
		end = &t
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

	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")

	const layout = "02-01-2006"

	var start *time.Time
	var end *time.Time

	if startStr != "" {
		t, err := time.Parse(layout, startStr)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Invalid --start date format, expected DD-MM-YYYY, got %s", startStr)
		}
		start = &t
	}

	if endStr != "" {
		t, err := time.Parse(layout, endStr)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Invalid --end date format, expected DD-MM-YYYY, got %s", endStr)
		}
		// include full day
		t = t.AddDate(0, 0, 1)
		end = &t
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
		_ = json.NewEncoder(os.Stdout).Encode(records)
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
		inHuman := humanize.Bytes(r.In)
		outHuman := humanize.Bytes(r.Out)

		if hideDate || r.Timestamp == nil {
			fmt.Fprintf(
				w,
				"%s\t%d\t%s\t%s\t%s\n",
				r.ID,
				r.VMID,
				r.NodeID,
				inHuman,
				outHuman,
			)
			continue
		}

		displayDate := r.Timestamp.Format("02 Jan 2006")

		fmt.Fprintf(
			w,
			"%s\t%d\t%s\t%s\t%s\t%s\n",
			r.ID,
			r.VMID,
			r.NodeID,
			displayDate,
			inHuman,
			outHuman,
		)
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
