package hcmmigration

import (
	"bufio"
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/stefanistkuhl/ermokie/pkg/games"
	"github.com/stefanistkuhl/ermokie/pkg/games/categories"
)

func LoadProfiles(fileName string) (Profiles, error) {
	var xmlContents []string
	var profiles Profiles
	file, err := os.Open(fileName)
	if err != nil {
		return profiles, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	foundProfileSection := false
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "")
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Map(func(r rune) rune {
			if unicode.IsGraphic(r) {
				return r
			}
			return -1
		}, line)
		line = replacer.Replace(line)

		if line == "<Profiles>" {
			foundProfileSection = true
		}

		if foundProfileSection {
			xmlContents = append(xmlContents, line)
		}

		if line == "</Profiles>" {
			foundProfileSection = false
		}

	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	xmlData := strings.Join(xmlContents, "\n")
	err = xml.Unmarshal([]byte(xmlData), &profiles)
	if err != nil {
		panic(err)
	}
	return profiles, err
}

func ImportProfiles(db *sql.DB, profiles Profiles) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, profile := range profiles.ProfileList.Profiles {
		cat := games.NewCatalog()
		catCatalog := categories.NewCatalog()
		g, ok := games.ParseProfileGame(cat, profile.Name)
		var gameDB any
		if ok {
			gameDB = g.Name
		} else {
			gameDB = nil
		}
		c, ok := categories.ParseProfileCategory(catCatalog, profile.Name)
		var categoryDB any
		if ok {
			categoryDB = c.Name
		} else {
			categoryDB = nil
		}
		res, err := tx.Exec(
			"INSERT INTO runs (name, attempts, active_split, game, category) VALUES (?, ?, ?, ?, ?)",
			profile.Name, profile.Attempts, profile.ActiveSplit, gameDB, categoryDB)
		if err != nil {
			return err
		}

		fmt.Println("Importing profile:", profile.Name)

		runID, _ := res.LastInsertId()

		for idx, row := range profile.Rows.ProfileRow {
			_, err := tx.Exec(
				"INSERT INTO splits (run_id, name, hit_count, pb_hit_count, idx) VALUES (?, ?, ?, ?, ?)",
				runID, row.Title, row.Hits, row.PB, idx+1,
			)
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
