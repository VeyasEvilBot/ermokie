package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/ermokie/pkg/db"
	hcmmigration "github.com/stefanistkuhl/ermokie/pkg/hcmMigration"
)

var importCmd = &cobra.Command{
	Use:   "import [filename]",
	Short: "Import HCM profiles from an XML file",
	Long:  `Import HCM profiles from the specified XML file into the database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		store, getDBErr := db.Init()
		if getDBErr != nil {
			fmt.Printf("Error getting database: %v\n", getDBErr)
			os.Exit(1)
		}

		if !db.CheckEmpty(store.DB) {
			fmt.Println("Database already contains data. Skipping import.")
			return
		}

		profiles, err := hcmmigration.LoadProfiles(filename)
		if err != nil {
			fmt.Printf("Error loading profiles: %v\n", err)
			os.Exit(1)
		}

		err = hcmmigration.ImportProfiles(store, profiles)
		if err != nil {
			fmt.Printf("Error importing profiles: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Profiles imported successfully!")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
