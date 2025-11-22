package hcmmigration

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/xml"
	"os"
	"strings"
	"unicode"

	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/db/sqlc"
	"codeberg.org/veya/ermokie/pkg/games"
	"codeberg.org/veya/ermokie/pkg/games/categories"
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

func ImportProfiles(s *db.Store, profiles Profiles) error {
	ctx := context.Background()
	tx, err := s.DB.BeginTx(ctx, nil)
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

		insProfile := sqlc.InsertRunParams{
			Name:        profile.Name,
			Attempts:    sql.NullInt64{Int64: int64(profile.Attempts), Valid: true},
			ActiveSplit: sql.NullInt64{Int64: int64(profile.ActiveSplit), Valid: true},
			Game:        sql.NullString{String: gameDB.(string), Valid: true},
			Category:    sql.NullString{String: categoryDB.(string), Valid: true},
		}
		id, err := s.Queries.InsertRun(context.Background(), insProfile)
		if err != nil {
			return err
		}

		for idx, row := range profile.Rows.ProfileRow {
			splitIns := sqlc.InsertSplitParams{
				RunID:      id,
				Name:       row.Title,
				HitCount:   sql.NullInt64{Int64: int64(row.Hits), Valid: true},
				PbHitCount: sql.NullInt64{Int64: int64(row.PB), Valid: true},
				Idx:        int64(idx + 1),
				SaveFile:   sql.NullString{String: "", Valid: false},
			}
			err := s.Queries.InsertSplit(context.Background(), splitIns)

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
