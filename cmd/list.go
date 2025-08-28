package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/ermokie/pkg/db"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List data from the database",
	Long:  `List various types of data stored in the database.`,
}

var listProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List all profiles/runs",
	Long:  `List all profiles and runs stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		database := db.Init()

		runs, err := db.GetAllRuns(database)
		if err != nil {
			fmt.Printf("Error getting runs: %v\n", err)
			os.Exit(1)
		}

		if len(runs) == 0 {
			fmt.Println("No profiles found in database.")
			return
		}

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

func init() {
	listCmd.AddCommand(listProfilesCmd)

	rootCmd.AddCommand(listCmd)
}
