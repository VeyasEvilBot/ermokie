package cmd

import (
	"context"
	"fmt"
	"os"

	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/ipc"
	"codeberg.org/veya/ermokie/pkg/utils"
	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Manage splits",
	Long:  `Manage splits for runs.`,
}

var splitNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Advance to the next split",
	Long:  `Advance to the next split in the active run.`,
	Run: func(cmd *cobra.Command, args []string) {
		type rawResult struct {
			Error       string
			ID          int
			Name        string
			Game        string
			Category    string
			Attempts    int
			ActiveSplit int
			NumSplits   int
		}
		rawRes := rawResult{}
		ctx := context.Background()

		rawFlag, _ := cmd.Flags().GetBool("raw")
		store, getDBErr := db.Init()
		if getDBErr != nil {
			if rawFlag {
				rawRes.Error = getDBErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error opening database: %v\n", getDBErr)
			}
			os.Exit(1)
		}

		tx, startTxErr := store.DB.BeginTx(ctx, nil)
		if startTxErr != nil {
			if rawFlag {
				rawRes.Error = startTxErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error starting transaction: %v\n", startTxErr)
			}
			os.Exit(1)
		}
		qtx := store.Queries.WithTx(tx)

		run, getRunErr := qtx.GetActiveRun(ctx)

		if getRunErr != nil {
			if rawFlag {
				rawRes.Error = getRunErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
				os.Exit(1)
			} else {
				fmt.Printf("Error getting active run: %v\n", getRunErr)
			}
			tx.Rollback()
			os.Exit(1)
		}

		splits, getSplitsErr := qtx.GetSplitsByRunID(ctx, run.ID)

		if getSplitsErr != nil {
			if rawFlag {
				rawRes.Error = getSplitsErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error getting splits: %v\n", getSplitsErr)
			}
			os.Exit(1)
		}

		rawRes.ID = int(run.ID)
		rawRes.Name = run.Name
		rawRes.Game = run.Game.String
		rawRes.Category = run.Category.String
		rawRes.Attempts = int(run.Attempts.Int64)
		rawRes.ActiveSplit = int(run.ActiveSplit.Int64)
		rawRes.NumSplits = len(splits)

		if run.ActiveSplit.Int64+1 > int64(len(splits)) {
			if rawFlag {
				rawRes.Error = "Error: No more splits to advance to"
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Println("Error: No more splits to advance to")
			}
			tx.Rollback()
			os.Exit(1)
		}

		advanceErr := qtx.AdvanceSplitInActiveRun(ctx)
		if advanceErr != nil {
			if rawFlag {
				rawRes.Error = advanceErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error advancing split: %v\n", advanceErr)
			}
			tx.Rollback()
			os.Exit(1)
		}

		commitErr := tx.Commit()
		if commitErr != nil {
			if rawFlag {
				rawRes.Error = commitErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error committing transaction: %v\n", commitErr)
			}
			os.Exit(1)
		}

		rawRes.ActiveSplit = int(run.ActiveSplit.Int64) + 1

		client, getClientErr := ipc.GetIPCClient()
		if getClientErr != nil {
			if rawFlag {
				rawRes.Error = getClientErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
				os.Exit(1)
			}
			os.Exit(1)
		}

		_, err := client.HandleRequest(ctx, &ipc.Request{
			Id: "1",
			Payload: &ipc.Request_AdvanceSplit{
				AdvanceSplit: &ipc.AdvanceSplit{
					RunId:          int64(rawRes.ID),
					ActiveSplitIdx: int64(rawRes.ActiveSplit),
				},
			},
		})

		if err != nil {
			if rawFlag {
				rawRes.Error = err.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
				os.Exit(1)
			}
		}
		if rawFlag {
			utils.PrintRawJSON(rawRes, rawFlag)
		}
	},
}

var splitPrevCmd = &cobra.Command{
	Use:   "prev",
	Short: "Go to the previous split",
	Long:  `Go to the previous split in the active run.`,
	Run: func(cmd *cobra.Command, args []string) {
		type rawResult struct {
			Error       string
			ID          int
			Name        string
			Game        string
			Category    string
			Attempts    int
			ActiveSplit int
			NumSplits   int
		}
		rawRes := rawResult{}

		rawFlag, _ := cmd.Flags().GetBool("raw")
		store, getDBErr := db.Init()
		if getDBErr != nil {
			if rawFlag {
				rawRes.Error = getDBErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error opening database: %v\n", getDBErr)
			}
			os.Exit(1)
		}

		ctx := context.Background()
		tx, startTxErr := store.DB.BeginTx(ctx, nil)

		if startTxErr != nil {
			if rawFlag {
				rawRes.Error = startTxErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error starting transaction: %v\n", startTxErr)
			}
			os.Exit(1)
		}

		qtx := store.Queries.WithTx(tx)

		run, getRunErr := qtx.GetActiveRun(ctx)

		if getRunErr != nil {
			if rawFlag {
				rawRes.Error = getRunErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
				os.Exit(1)
			} else {
				fmt.Printf("Error getting active run: %v\n", getRunErr)
			}
			tx.Rollback()
			os.Exit(1)
		}

		splits, getSplitsErr := qtx.GetSplitsByRunID(ctx, run.ID)

		if getSplitsErr != nil {
			if rawFlag {
				rawRes.Error = getSplitsErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error getting splits: %v\n", getSplitsErr)
			}
			os.Exit(1)
		}

		rawRes.ID = int(run.ID)
		rawRes.Name = run.Name
		rawRes.Game = run.Game.String
		rawRes.Category = run.Category.String
		rawRes.Attempts = int(run.Attempts.Int64)
		rawRes.ActiveSplit = int(run.ActiveSplit.Int64)
		rawRes.NumSplits = len(splits)

		if run.ActiveSplit.Int64-1 < 1 {
			if rawFlag {
				rawRes.Error = "Error: No more splits to go back to"
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Println("Error: No more splits to go back to")
			}
			tx.Rollback()
			os.Exit(1)
		}

		returnSplitErr := qtx.GoBackSplitInActiveRun(ctx)
		if returnSplitErr != nil {
			if rawFlag {
				rawRes.Error = returnSplitErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error advancing going back a split: %v\n", returnSplitErr)
			}
			tx.Rollback()
			os.Exit(1)
		}

		commitErr := tx.Commit()
		if commitErr != nil {
			if rawFlag {
				rawRes.Error = commitErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error committing transaction: %v\n", commitErr)
			}
			os.Exit(1)
		}

		rawRes.ActiveSplit = int(run.ActiveSplit.Int64) - 1

		client, getClientErr := ipc.GetIPCClient()
		if getClientErr != nil {
			if rawFlag {
				rawRes.Error = getClientErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
				os.Exit(1)
			}
			os.Exit(1)
		}
		_, err := client.HandleRequest(ctx, &ipc.Request{
			Id: "1",
			Payload: &ipc.Request_MoveActiveSplitBack{
				MoveActiveSplitBack: &ipc.MoveActiveSplitBack{
					RunId:          int64(rawRes.ID),
					ActiveSplitIdx: int64(rawRes.ActiveSplit),
				},
			},
		})

		if err != nil {
			if rawFlag {
				rawRes.Error = err.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
				os.Exit(1)
			}
		}
		if rawFlag {
			utils.PrintRawJSON(rawRes, rawFlag)
		}
	},
}

func init() {
	splitNextCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")
	splitPrevCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")

	splitCmd.AddCommand(splitNextCmd)
	splitCmd.AddCommand(splitPrevCmd)
	rootCmd.AddCommand(splitCmd)
}
