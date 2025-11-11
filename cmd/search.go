package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/ermokie/pkg/db"
	"github.com/stefanistkuhl/ermokie/pkg/types"
	"github.com/stefanistkuhl/ermokie/pkg/utils"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search and filter runs",
	Long:  `Search and filter runs by various criteria like game, category, or name.`,
}

var searchRunsCmd = &cobra.Command{
	Use:   "runs [query]",
	Short: "Search runs by name, game, or category",
	Long:  `Search runs by name (partial match), game, or category.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rawFlag, _ := cmd.Flags().GetBool("raw")
		query := strings.ToLower(args[0])
		store, getDBErr := db.Init()
		if getDBErr != nil {
			fmt.Printf("Error getting database: %v\n", getDBErr)
			os.Exit(1)
		}

		allRuns, err := store.GetAllRuns(context.Background())
		if err != nil {
			fmt.Printf("Error getting runs: %v\n", err)
			os.Exit(1)
		}

		var filteredRuns []types.Run
		for _, run := range allRuns {
			if strings.Contains(strings.ToLower(run.Name), query) {
				filteredRuns = append(filteredRuns, types.Run{
					ID:          int(run.ID),
					Name:        run.Name,
					Game:        run.Game,
					Category:    run.Category,
					Attempts:    int(run.Attempts.Int64),
					ActiveSplit: int(run.ActiveSplit.Int64),
				})
				continue
			}

			if run.Game.Valid && strings.Contains(strings.ToLower(run.Game.String), query) {
				filteredRuns = append(filteredRuns, types.Run{
					ID:          int(run.ID),
					Name:        run.Name,
					Game:        run.Game,
					Category:    run.Category,
					Attempts:    int(run.Attempts.Int64),
					ActiveSplit: int(run.ActiveSplit.Int64),
				})
				continue
			}

			if run.Category.Valid && strings.Contains(strings.ToLower(run.Category.String), query) {
				filteredRuns = append(filteredRuns, types.Run{
					ID:          int(run.ID),
					Name:        run.Name,
					Game:        run.Game,
					Category:    run.Category,
					Attempts:    int(run.Attempts.Int64),
					ActiveSplit: int(run.ActiveSplit.Int64),
				})
				continue
			}
		}

		if len(filteredRuns) == 0 {
			fmt.Printf("No runs found matching: %s\n", query)
			return
		}

		utils.PrintRawJSON(filteredRuns, rawFlag)

		fmt.Printf("Found %d runs matching '%s':\n\n", len(filteredRuns), query)

		maxID := 2
		maxName := 4
		maxGame := 4
		maxCategory := 8
		maxAttempts := 8

		for _, run := range filteredRuns {
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

		fmt.Printf("%-*s  %-*s  %-*s  %-*s  %-*s\n",
			maxID, "ID", maxName, "Name", maxGame, "Game", maxCategory, "Category", maxAttempts, "Attempts")

		separator := strings.Repeat("-", maxID) + "  " + strings.Repeat("-", maxName) + "  " +
			strings.Repeat("-", maxGame) + "  " + strings.Repeat("-", maxCategory) + "  " + strings.Repeat("-", maxAttempts)
		fmt.Println(separator)

		for _, run := range filteredRuns {
			game := "N/A"
			if run.Game.Valid {
				game = run.Game.String
			}
			category := "N/A"
			if run.Category.Valid {
				category = run.Category.String
			}
			fmt.Printf("%-*d  %-*s  %-*s  %-*s  %-*d\n",
				maxID, run.ID, maxName, run.Name, maxGame, game, maxCategory, category, maxAttempts, run.Attempts)
		}
	},
}

func init() {
	searchRunsCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")

	searchCmd.AddCommand(searchRunsCmd)
	rootCmd.AddCommand(searchCmd)
}
