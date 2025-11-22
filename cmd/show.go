package cmd

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/db/sqlc"
	"codeberg.org/veya/ermokie/pkg/utils"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show detailed information",
	Long:  `Show detailed information about various entities.`,
}

type Run = sqlc.GetRunByIDRow
type Split = sqlc.GetSplitsByRunIDRow

var showRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Show all runs",
	Long:  `Show all runs stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		rawFlag, _ := cmd.Flags().GetBool("raw")
		store, getDbErr := db.Init()
		if getDbErr != nil {
			fmt.Printf("Failed to initialize the database %s", getDbErr)
			return
		}

		runs, err := store.GetAllRuns(context.Background())
		if err != nil {
			fmt.Printf("Error getting runs: %v\n", err)
			os.Exit(1)
		}

		if len(runs) == 0 {
			fmt.Println("No runs found in database.")
			return
		}

		utils.PrintRawJSON(runs, rawFlag)
		os.Exit(0)

		maxID := 2
		maxName := 4
		maxGame := 4
		maxCategory := 8
		maxAttempts := 8

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
		store, getDbErr := db.Init()
		if getDbErr != nil {
			fmt.Printf("Failed to initialize the database %s", getDbErr)
			return
		}

		var run Run

		runID, parseErr := strconv.Atoi(runArg)
		if parseErr == nil {
			runTmp, err := store.GetRunByID(context.Background(), int64(runID))
			if err != nil {
				fmt.Printf("Error getting run info: %v\n", err)
				os.Exit(1)
			}
			run = runTmp
		} else {
			runTmp, err := store.GetRunByName(context.Background(), runArg)
			if err != nil {
				fmt.Printf("Error getting run info: %v\n", err)
				os.Exit(1)
			}
			run = Run(runTmp)
		}

		splits, err := store.GetSplitsByRunID(context.Background(), run.ID)
		if err != nil {
			fmt.Printf("Error getting splits: %v\n", err)
			os.Exit(1)
		}

		if rawFlag {
			runWithSplits := struct {
				Run    Run     `json:"run"`
				Splits []Split `json:"splits"`
			}{
				Run:    run,
				Splits: splits,
			}
			utils.PrintRawJSON(runWithSplits, rawFlag)
			os.Exit(0)
		}

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

		if len(splits) > 0 {
			fmt.Println("=== Splits Summary ===")
			totalHits := 0
			totalPBHits := 0

			for _, split := range splits {
				totalHits += int(split.HitCount.Int64)
				totalPBHits += int(split.PbHitCount.Int64)
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
		store, getDbErr := db.Init()
		if getDbErr != nil {
			fmt.Printf("Failed to initialize the database %s", getDbErr)
			return
		}

		var splits []Split
		var run Run

		runID, parseErr := strconv.Atoi(runArg)
		if parseErr == nil {
			splitsTmp, err := store.GetSplitsByRunID(context.Background(), int64(runID))
			if err != nil {
				fmt.Printf("Error getting splits: %v\n", err)
				os.Exit(1)
			}
			splits = splitsTmp
			runTmp, err := store.GetRunByID(context.Background(), int64(runID))
			if err != nil {
				fmt.Printf("Error getting run info: %v\n", err)
				os.Exit(1)
			}
			run = runTmp
		} else {
			splitsTmp, err := store.GetSplitsByRunName(context.Background(), runArg)
			if err != nil {
				fmt.Printf("Error getting splits: %v\n", err)
				os.Exit(1)
			}
			for _, split := range splitsTmp {
				splits = append(splits, Split(split))
			}
			runTmp, err := store.GetRunByName(context.Background(), runArg)
			if err != nil {
				fmt.Printf("Error getting run info: %v\n", err)
				os.Exit(1)
			}
			run = Run(runTmp)
		}

		if len(splits) == 0 {
			fmt.Printf("No splits found for run: %s\n", runArg)
			return
		}

		if rawFlag {
			runWithSplits := struct {
				Run    Run     `json:"run"`
				Splits []Split `json:"splits"`
			}{
				Run:    run,
				Splits: splits,
			}
			utils.PrintRawJSON(runWithSplits, rawFlag)
			os.Exit(0)
		}

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

		maxIdx := 3
		maxName := 4
		maxHits := 4
		maxPBHits := 7
		maxSaveFile := 9
		maxActive := 6

		for _, split := range splits {
			if len(fmt.Sprintf("%d", split.Idx)) > maxIdx {
				maxIdx = len(fmt.Sprintf("%d", split.Idx))
			}
			if len(split.Name) > maxName {
				maxName = len(split.Name)
			}
			if len(fmt.Sprintf("%d", split.HitCount)) > maxHits {
				maxHits = len(fmt.Sprintf("%d", split.HitCount))
			}
			if len(fmt.Sprintf("%d", split.PbHitCount.Int64)) > maxPBHits {
				maxPBHits = len(fmt.Sprintf("%d", split.PbHitCount.Int64))
			}
			saveFile := "N/A"
			if split.SaveFile.Valid {
				saveFile = split.SaveFile.String
			}
			if len(saveFile) > maxSaveFile {
				maxSaveFile = len(saveFile)
			}
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s  %-*s  %-*s\n",
			maxIdx, "Idx", maxName, "Name", maxHits, "Hits", maxPBHits, "PB Hits", maxSaveFile, "Save File", maxActive, "Active")

		separator := strings.Repeat("-", maxIdx) + "  " + strings.Repeat("-", maxName) + "  " +
			strings.Repeat("-", maxHits) + "  " + strings.Repeat("-", maxPBHits) + "  " + strings.Repeat("-", maxSaveFile) + "  " + strings.Repeat("-", maxActive)
		fmt.Println(separator)

		for _, split := range splits {
			saveFile := "N/A"
			if split.SaveFile.Valid {
				saveFile = split.SaveFile.String
			}
			active := "No"
			isActive := reflect.ValueOf(split.IsActive).Bool()
			if isActive {
				active = "Yes"
			}
			fmt.Printf("%-*d  %-*s  %-*d  %-*d  %-*s  %-*s\n",
				maxIdx, split.Idx, maxName, split.Name, maxHits, split.HitCount.Int64, maxPBHits, split.PbHitCount.Int64, maxSaveFile, saveFile, maxActive, active)
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
		store, getDbErr := db.Init()
		if getDbErr != nil {
			fmt.Printf("Failed to initialize the database %s", getDbErr)
			return
		}

		var split sqlc.Split
		var run Run

		runID, parseErr := strconv.Atoi(runArg)
		if parseErr == nil {
			var err error
			split, err = store.GetSplitByRunIDAndIdx(context.Background(), sqlc.GetSplitByRunIDAndIdxParams{RunID: int64(runID), Idx: int64(splitIdx)})
			if err != nil {
				fmt.Printf("Error getting split info: %v\n", err)
				os.Exit(1)
			}
			runTmp, err := store.GetRunByID(context.Background(), int64(runID))
			if err != nil {
				fmt.Printf("Error getting run info: %v\n", err)
				os.Exit(1)
			}
			run = Run(runTmp)
		} else {
			var err error
			split, err = store.GetSplitByRunNameAndIdx(context.Background(), sqlc.GetSplitByRunNameAndIdxParams{Name: runArg, Idx: int64(splitIdx)})
			if err != nil {
				fmt.Printf("Error getting split info: %v\n", err)
				os.Exit(1)
			}
			runTmp, err := store.GetRunByName(context.Background(), runArg)
			if err != nil {
				fmt.Printf("Error getting run info: %v\n", err)
				os.Exit(1)
			}
			run = Run(runTmp)
		}

		if rawFlag {
			splitWithRun := struct {
				Run   Run        `json:"run"`
				Split sqlc.Split `json:"split"`
			}{
				Run:   run,
				Split: split,
			}
			utils.PrintRawJSON(splitWithRun, rawFlag)
			os.Exit(0)
		}

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
		fmt.Printf("Hits: %d\n", split.HitCount.Int64)
		fmt.Printf("PB Hits: %d\n", split.PbHitCount.Int64)
		fmt.Printf("Active: %s\n", func() string {
			if run.ActiveSplit.Int64 == split.Idx {
				return "Yes"
			}
			return "No"
		})

		if split.SaveFile.Valid {
			fmt.Printf("Save File: %s\n", split.SaveFile.String)
		} else {
			fmt.Println("Save File: N/A")
		}
	},
}

func init() {
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
