package rendering

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"path/filepath"

	overlaythemes "codeberg.org/veya/ermokie/pkg/rendering/overlayThemes"
	"codeberg.org/veya/ermokie/pkg/types"
)

type TemplateName string
type TemplateCategory string

const (
	TemplateNameBase TemplateName = "base"
)
const (
	TemplateCategoryBase TemplateCategory = "base"
)

type PageData struct {
	RunName string
	Splits  []SplitData
	CssName string
}

type SplitData struct {
	SplitName string
	Hits      int
	PB        int
	Diff      int
	IsActive  bool
}

type DisplayCountSettingsData struct {
	IsLimited bool
	Count     int
}

type CounterCountSettings struct {
	DisplayPrev DisplayCountSettingsData
	DisplayNext DisplayCountSettingsData
}

type RenderingParams struct {
	Run             types.Run
	Splits          []types.Split
	Tc              TemplateCategory
	Tn              TemplateName
	CssFileName     string
	CounterSettings CounterCountSettings
	Theme           overlaythemes.OverlayTheme
}

//go:embed template/*
var templateFS embed.FS

func RenderHtml(params RenderingParams) ([]byte, []byte, error) {
	if params.Tc != TemplateCategoryBase {
		return nil, nil, fmt.Errorf("invalid template category: %s", params.Tc)
	}
	if params.Tn != TemplateNameBase {
		return nil, nil, fmt.Errorf("invalid template name: %s", params.Tn)
	}
	templ, err := template.ParseFS(templateFS, filepath.Join("template", string(params.Tc), string(params.Tn)+".html"))
	if err != nil {
		return nil, nil, err
	}

	data := PageData{
		RunName: params.Run.Name,
		Splits:  make([]SplitData, len(params.Splits)),
		CssName: params.CssFileName,
	}

	activeSplit := 0
	for i, split := range params.Splits {
		if split.IsActive {
			activeSplit = i
			break
		}
	}

	data.Splits = []SplitData{}
	for i, split := range params.Splits {
		if params.CounterSettings.DisplayNext.IsLimited &&
			i < activeSplit-params.CounterSettings.DisplayNext.Count {
			continue
		}
		if params.CounterSettings.DisplayPrev.IsLimited &&
			i > activeSplit+params.CounterSettings.DisplayPrev.Count {
			continue
		}
		data.Splits = append(data.Splits, SplitData{SplitName: split.Name,
			Hits:     split.Hits,
			PB:       split.PBHits,
			Diff:     split.Diff,
			IsActive: split.IsActive})
	}

	var buf bytes.Buffer
	var bufCss bytes.Buffer
	cssFile, err := templateFS.Open(filepath.Join("template", string(params.Tc), string(params.Tn)+".css"))
	if err != nil {
		return nil, nil, err
	}
	defer cssFile.Close()
	cssContent, getCssContentErr := io.ReadAll(cssFile)
	if getCssContentErr != nil {
		return nil, nil, getCssContentErr
	}
	cssTempl, createCssTmplErr := template.New("css").Delims("@@", "@@").Parse(string(cssContent))
	if createCssTmplErr != nil {
		return nil, nil, createCssTmplErr
	}
	err = cssTempl.Execute(&bufCss, params.Theme.Colors)
	if err != nil {
		return nil, nil, err
	}

	err = templ.Execute(&buf, data)

	return buf.Bytes(), bufCss.Bytes(), nil
}
