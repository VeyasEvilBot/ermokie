package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/pretty"
)

// PrintRawJSON prints the data structure as formatted JSON if --raw flag is set
// Usage: Add this to any command that needs --raw JSON output
func PrintRawJSON(data interface{}, rawFlag bool) {
	if !rawFlag {
		return
	}

	// Marshal to JSON first
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Use tidwall/pretty for nice formatting
	prettyData := pretty.Pretty(jsonData)
	prettyData = pretty.Color(prettyData, nil)
	fmt.Println(string(prettyData))
	os.Exit(0)
}
