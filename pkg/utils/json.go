package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/pretty"
)

func PrintRawJSON(data any, rawFlag bool) {
	if !rawFlag {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	prettyData := pretty.Pretty(jsonData)
	fmt.Println(string(prettyData))
}
