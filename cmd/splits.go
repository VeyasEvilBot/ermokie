package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"

	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/globals"
	"codeberg.org/veya/ermokie/pkg/ipc"
	"codeberg.org/veya/ermokie/pkg/utils"
	"github.com/spf13/cobra"
	"storj.io/drpc/drpcconn"
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

		tx, startTxErr := store.DB.BeginTx(context.Background(), nil)

		if startTxErr != nil {
			if rawFlag {
				rawRes.Error = startTxErr.Error()
				utils.PrintRawJSON(rawRes, rawFlag)
			} else {
				fmt.Printf("Error starting transaction: %v\n", startTxErr)
			}
			os.Exit(1)
		}

		run, getRunErr := store.Queries.GetActiveRun(context.Background())

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

		splits, getSplitsErr := store.Queries.GetSplitsByRunID(context.Background(), run.ID)

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

		advanceErr := store.Queries.AdvanceSplitInActiveRun(context.Background())
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

		if runtime.GOOS == "windows" {
			rawRes.Error = "Error: IPC not implemented on Windows yet"
			if rawFlag {
				utils.PrintRawJSON(rawRes, rawFlag)
				return
			}
		} else {
			rawConn, rawConnErr := net.Dial("unix", globals.UnixIPCSocketPath)
			if rawConnErr != nil {
				if rawFlag {
					rawRes.Error = rawConnErr.Error()
					utils.PrintRawJSON(rawRes, rawFlag)
					os.Exit(1)
				}
			}
			conn := drpcconn.New(rawConn)
			defer conn.Close()

			client := ipc.NewDRPCIPCClient(conn)

			_, err := client.HandleRequest(context.Background(), &ipc.Request{
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
		}
		if rawFlag {
			utils.PrintRawJSON(rawRes, rawFlag)
		}
	},
}

func init() {
	splitNextCmd.Flags().BoolP("raw", "r", false, "Output raw JSON data")

	splitCmd.AddCommand(splitNextCmd)
	rootCmd.AddCommand(splitCmd)
}
