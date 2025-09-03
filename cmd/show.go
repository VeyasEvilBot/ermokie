package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/ermokie/pkg/db"
	"github.com/stefanistkuhl/ermokie/pkg/db/fetching"
	"github.com/stefanistkuhl/ermokie/pkg/utils"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show detailed information",
	Long:  `Show detailed information about various entities.`,
}

var showRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Show all runs",
	Long:  `Show all runs stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		rawFlag, _ := cmd.Flags().GetBool("raw")
		database := db.Init()

		runs, err := fetching.GetAllRuns(database)
		if err != nil {
			fmt.Printf("Error getting runs: %v\n", err)
			os.Exit(1)
		}

		if len(runs) == 0 {
			fmt.Println("No runs found in database.")
			return
		}

		// Check for --raw flag first
		utils.PrintRawJSON(runs, rawFlag)

		// Calculate column widths
		maxID := 2       // "ID" header
		maxName := 4     // "Name" header
		maxGame := 4     // "Game" header
		maxCategory := 8 // "Category" header
		maxAttempts := 8 // "Attempts" header

		for _, run := range runs {
			idLen := len(fmt.Sprintf("%d", run.ID))
			if idLen > maxID {
				maxID = idLen
			}
			if len(run.Name) > maxName {
				maxName = len(run.Name)
			}
			game := "N/A"
			if run.Game.Valid {
				game = run.Game.String
			}
			if len(game) > maxGame {
				maxGame = len(game)
			}
			category := "N/A"
			if run.Category.Valid {
				category = run.Category.String
			}
			if len(category) > maxCategory {
				maxCategory = len(category)
			}
			attemptsLen := len(fmt.Sprintf("%d", run.Attempts))
			if attemptsLen > maxAttempts {
				maxAttempts = attemptsLen
			}
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s  %-*s\n", maxID, "ID", maxName, "Name", maxGame, "Game", maxCategory, "Category", maxAttempts, "Attempts")

		separator := strings.Repeat("-", maxID) + "  " + strings.Repeat("-", maxName) + "  " + strings.Repeat("-", maxGame) + "  " + strings.Repeat("-", maxCategory) + "  " + strings.Repeat("-", maxAttempts)
		fmt.Println(separator)

		for _, run := range runs {
			game := "N/A"
			if run.Game.Valid {
				game = run.Game.String
			}
			category := "N/A"
			if run.Category.Valid {
				category = run.Category.String
			}
			fmt.Printf("%-*d  %-*s  %-*s  %-*s  %-*d\n", maxID, run.ID, maxName, run.Name, maxGame, game, maxCategory, category, maxAttempts, run.Attempts)
		}
	},
}

var showRunCmd = &cobra.Command{
	Use:   "run [id/name]",
	Short: "Show detailed information about a specific run",
	Long:  `Show detailed information about a specific run, identified by ID or name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rawFlag, _ := cmd.Flags().GetBool("raw")
		runArg := args[0]
		database := db.Init()

		var run *fetching.Run
		var err error

		// Try to parse as ID first, then as name
		if runID, parseErr := strconv.Atoi(runArg); parseErr == nil {
			// It's a numeric ID
			run, err = fetching.GetRunByID(database, runID)
		} else {
			// It's a name
			run, err = fetching.GetRunByName(database, runArg)
		}

		if err != nil {
			fmt.Printf("Error getting run info: %v\n", err)
			os.Exit(1)
		}

		// Get splits for this run
		splits, err := fetching.GetSplitsByRunID(database, run.ID)
		if err != nil {
			fmt.Printf("Error getting splits: %v\n", err)
			os.Exit(1)
		}

		// Check for --raw flag first
		if rawFlag {
			// Create a combined structure for raw output
			runWithSplits := struct {
				Run    *fetching.Run    `json:"run"`
				Splits []fetching.Split `json:"splits"`
			}{
				Run:    run,
				Splits: splits,
			}
			utils.PrintRawJSON(runWithSplits, rawFlag)
		}

		// Display run information
		fmt.Println("=== Run Information ===")
		fmt.Printf("ID: %d\n", run.ID)
		fmt.Printf("Name: %s\n", run.Name)

		if run.Game.Valid {
			fmt.Printf("Game: %s\n", run.Game.String)
		} else {
			fmt.Println("Game: N/A")
		}

		if run.Category.Valid {
			fmt.Printf("Category: %s\n", run.Category.String)
		} else {
			fmt.Println("Category: N/A")
		}

		fmt.Printf("Attempts: %d\n", run.Attempts)
		fmt.Printf("Active Split: %d\n", run.ActiveSplit)
		fmt.Printf("Total Splits: %d\n\n", len(splits))

		// Display splits summary
		if len(splits) > 0 {
			fmt.Println("=== Splits Summary ===")
			totalHits := 0
			totalPBHits := 0

			for _, split := range splits {
				totalHits += split.Hits
				totalPBHits += split.PBHits
			}

			fmt.Printf("Total Hits: %d\n", totalHits)
			fmt.Printf("Total PB Hits: %d\n", totalPBHits)

			if totalPBHits > 0 {
				improvement := totalHits - totalPBHits
				if improvement > 0 {
					fmt.Printf("Improvement Potential: +%d hits\n", improvement)
				} else if improvement < 0 {
					fmt.Printf("Current Best: %d hits better than PB\n", -improvement)
				} else {
					fmt.Println("Current Run: Matches PB exactly!")
				}
			}
		} else {
			fmt.Println("No splits found for this run.")
		}
	},
}

var showSplitsCmd = &cobra.Command{
	Use:   "splits [id/name]",
	Short: "Show splits for a specific run",
	Long:  `Show all splits for a specific run, identified by ID or name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rawFlag, _ := cmd.Flags().GetBool("raw")
		runArg := args[0]
		database := db.Init()

		var splits []fetching.Split
		var run *fetching.Run
		var err error

		// Try to parse as ID first, then as name
		if runID, parseErr := strconv.Atoi(runArg); parseErr == nil {
			// It's a numeric ID
			splits, err = fetching.GetSplitsByRunID(database, runID)
			if err != nil {
				fmt.Printf("Error getting splits: %v\n", err)
				os.Exit(1)
			}
			run, err = fetching.GetRunByID(database, runID)
		} else {
			// It's a name
			splits, err = fetching.GetSplitsByRunName(database, runArg)
			if err != nil {
				fmt.Printf("Error getting splits: %v\n", err)
				os.Exit(1)
			}
			run, err = fetching.GetRunByName(database, runArg)
		}

		if err != nil {
			fmt.Printf("Error getting run info: %v\n", err)
			os.Exit(1)
		}

		if len(splits) == 0 {
			fmt.Printf("No splits found for run: %s\n", runArg)
			return
		}

		// Check for --raw flag first
		if rawFlag {
			// Create a combined structure for raw output
			runWithSplits := struct {
				Run    *fetching.Run    `json:"run"`
				Splits []fetching.Split `json:"splits"`
			}{
				Run:    run,
				Splits: splits,
			}
			utils.PrintRawJSON(runWithSplits, rawFlag)
		}

		// Display run information
		fmt.Printf("Run: %s\n", run.Name)
		if run.Game.Valid {
			fmt.Printf("Game: %s\n", run.Game.String)
		}
		if run.Category.Valid {
			fmt.Printf("Category: %s\n", run.Category.String)
		}
		fmt.Printf("Attempts: %d\n", run.Attempts)
		fmt.Printf("Active Split: %d\n", run.ActiveSplit)
		fmt.Printf("Total Splits: %d\n\n", len(splits))

		// Calculate column widths for splits table
		maxIdx := 3      // "Idx" header
		maxName := 4     // "Name" header
		maxHits := 4     // "Hits" header
		maxPBHits := 7   // "PB Hits" header
		maxSaveFile := 9 // "Save File" header
		maxActive := 6   // "Active" header

		for _, split := range splits {
			if len(fmt.Sprintf("%d", split.Idx)) > maxIdx {
				maxIdx = len(fmt.Sprintf("%d", split.Idx))
			}
			if len(split.Name) > maxName {
				maxName = len(split.Name)
			}
			if len(fmt.Sprintf("%d", split.Hits)) > maxHits {
				maxHits = len(fmt.Sprintf("%d", split.Hits))
			}
			if len(fmt.Sprintf("%d", split.PBHits)) > maxPBHits {
				maxPBHits = len(fmt.Sprintf("%d", split.PBHits))
			}
			saveFile := "N/A"
			if split.SaveFile.Valid {
				saveFile = split.SaveFile.String
			}
			if len(saveFile) > maxSaveFile {
				maxSaveFile = len(saveFile)
			}
		}

		// Print splits table header
		fmt.Printf("%-*s  %-*s  %-*s  %-*s  %-*s  %-*s\n",
			maxIdx, "Idx", maxName, "Name", maxHits, "Hits", maxPBHits, "PB Hits", maxSaveFile, "Save File", maxActive, "Active")

		// Print separator line
		separator := strings.Repeat("-", maxIdx) + "  " + strings.Repeat("-", maxName) + "  " +
			strings.Repeat("-", maxHits) + "  " + strings.Repeat("-", maxPBHits) + "  " + strings.Repeat("-", maxSaveFile) + "  " + strings.Repeat("-", maxActive)
		fmt.Println(separator)

		// Print splits data
		for _, split := range splits {
			saveFile := "N/A"
			if split.SaveFile.Valid {
				saveFile = split.SaveFile.String
			}
			active := "No"
			if split.IsActive {
				active = "Yes"
			}
			fmt.Printf("%-*d  %-*s  %-*d  %-*d  %-*s  %-*s\n",
				maxIdx, split.Idx, maxName, split.Name, maxHits, split.Hits, maxPBHits, split.PBHits, maxSaveFile, saveFile, maxActive, active)
		}
	},
}

var showSplitCmd = &cobra.Command{
	Use:   "split [id/name] [idx]",
	Short: "Show detailed information about a specific split",
	Long:  `Show detailed information about a specific split, identified by run ID/name and split index.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		rawFlag, _ := cmd.Flags().GetBool("raw")
		runArg := args[0]
		splitIdx, _ := strconv.Atoi(args[1])
		database := db.Init()

		var split *fetching.Split
		var run *fetching.Run
		var err error

		// Try to parse as ID first, then as name
		if runID, parseErr := strconv.Atoi(runArg); parseErr == nil {
			// It's a numeric ID
			split, err = fetching.GetSplitByRunIDAndIdx(database, runID, splitIdx)
			run, err = fetching.GetRunByID(database, runID)
		} else {
			// It's a name
			split, err = fetching.GetSplitByRunNameAndIdx(database, runArg, splitIdx)
			run, err = fetching.GetRunByName(database, runArg)
		}

		if err != nil {
			fmt.Printf("Error getting split info: %v\n", err)
			os.Exit(1)
		}

		// Check for --raw flag first
		if rawFlag {
			// Create a combined structure for raw output
			splitWithRun := struct {
				Run   *fetching.Run   `json:"run"`
				Split *fetching.Split `json:"split"`
			}{
				Run:   run,
				Split: split,
			}
			utils.PrintRawJSON(splitWithRun, rawFlag)
		}

		// Display split information
		fmt.Printf("Run: %s\n", run.Name)
		if run.Game.Valid {
			fmt.Printf("Game: %s\n", run.Game.String)
		}
		if run.Category.Valid {
			fmt.Printf("Category: %s\n", run.Category.String)
		}
		fmt.Printf("Attempts: %d\n", run.Attempts)
		fmt.Printf("Active Split: %d\n", run.ActiveSplit)
		fmt.Printf("Split Index: %d\n", split.Idx)
		fmt.Printf("Split Name: %s\n", split.Name)
		fmt.Printf("Hits: %d\n", split.Hits)
		fmt.Printf("PB Hits: %d\n", split.PBHits)
		fmt.Printf("Active: %s\n", func() string {
			if split.IsActive {
				return "Yes"
			}
			return "No"
		}())

		if split.SaveFile.Valid {
			fmt.Printf("Save File: %s\n", split.SaveFile.String)
		} else {
			fmt.Println("Save File: N/A")
		}
	},
}

func init() {
	// Add --raw flag to all commands
	showRunsCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")
	showRunCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")
	showSplitsCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")
	showSplitCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")

	showCmd.AddCommand(showRunsCmd)
	showCmd.AddCommand(showRunCmd)
	showCmd.AddCommand(showSplitsCmd)
	showCmd.AddCommand(showSplitCmd)
	rootCmd.AddCommand(showCmd)
}
