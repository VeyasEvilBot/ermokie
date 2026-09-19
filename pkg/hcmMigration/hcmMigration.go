package hcmmigration

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/db/sqlc"
	"codeberg.org/veya/ermokie/pkg/games"
	"codeberg.org/veya/ermokie/pkg/games/categories"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func LoadProfiles(fileName string) (Profiles, error) {
	var xmlContents []string
	var profiles Profiles
	file, err := os.Open(fileName)
	if err != nil {
		return profiles, err
	}
	defer func() { _ = file.Close() }()

	transformer := unicode.BOMOverride(unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder())
	reader := transform.NewReader(file, transformer)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)

	foundProfileSection := false
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "")
	for scanner.Scan() {
		line := scanner.Text()
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
		return profiles, fmt.Errorf("read HCM profiles: %w", err)
	}

	xmlData := strings.Join(xmlContents, "\n")
	err = xml.Unmarshal([]byte(xmlData), &profiles)
	if err != nil {
		return profiles, fmt.Errorf("parse HCM profiles: %w", err)
	}
	return profiles, nil
}

func ImportProfiles(s *db.Store, profiles Profiles) error {
	ctx := context.Background()
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()

	cat := games.NewCatalog()
	catCatalog := categories.NewCatalog()

	qtx := s.Queries.WithTx(tx)

	for _, profile := range profiles.ProfileList.Profiles {

		gameField := sql.NullString{Valid: false}
		if g, ok := games.ParseProfileGame(cat, profile.Name); ok {
			gameField = sql.NullString{String: g.Name, Valid: true}
		}

		categoryField := sql.NullString{Valid: false}
		if c, ok := categories.ParseProfileCategory(catCatalog, profile.Name); ok {
			categoryField = sql.NullString{String: c.Name, Valid: true}
		}

		insProfile := sqlc.InsertRunParams{
			Name:        profile.Name,
			Attempts:    sql.NullInt64{Int64: int64(profile.Attempts), Valid: true},
			ActiveSplit: sql.NullInt64{Int64: int64(profile.ActiveSplit), Valid: true},
			Game:        gameField,
			Category:    categoryField,
		}

		id, err := qtx.InsertRun(ctx, insProfile)
		if err != nil {
			return err
		}

		for idx, row := range profile.Rows.ProfileRow {
			splitIns := sqlc.InsertSplitParams{
				RunID:      id,
				Name:       row.Title,
				HitCount:   sql.NullInt64{Int64: int64(row.Hits), Valid: true},
				PbHitCount: sql.NullInt64{Int64: int64(row.PB), Valid: true},
				Idx:        int64(idx),
				SaveFile:   sql.NullString{String: "", Valid: false},
			}
			err := qtx.InsertSplit(ctx, splitIns)
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
