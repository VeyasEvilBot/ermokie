package rendering

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	texttemplate "text/template"

	overlaythemes "codeberg.org/veya/ermokie/pkg/rendering/overlayThemes"
	"codeberg.org/veya/ermokie/pkg/types"
)

type TemplateName string
type TemplateCategory string

const (
	TemplateNameBase     TemplateName     = "base"
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
	// TemplateDir optionally points to a user template root. Custom templates
	// live at <TemplateDir>/<category>/<name>.{html,css}.
	TemplateDir string
}

//go:embed template/*/*
var templateFS embed.FS

var templatePartPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// RenderHTML renders a matched HTML/CSS template pair. The built-in base
// template is always available; user templates are loaded only from
// TemplateDir after strict path-component validation.
//
//nolint:gocritic // RenderingParams is a public value-oriented options struct.
func RenderHTML(params RenderingParams) ([]byte, []byte, error) {
	category := string(params.Tc)
	name := string(params.Tn)
	if category == "" {
		category = string(TemplateCategoryBase)
	}
	if name == "" {
		name = string(TemplateNameBase)
	}
	if !templatePartPattern.MatchString(category) || !templatePartPattern.MatchString(name) {
		return nil, nil, fmt.Errorf("invalid template %q/%q", category, name)
	}

	htmlSource, cssSource, err := loadTemplatePair(params.TemplateDir, category, name)
	if err != nil {
		return nil, nil, err
	}

	page := buildPageData(&params)
	htmlTmpl, err := htmltemplate.New(name + ".html").Parse(string(htmlSource))
	if err != nil {
		return nil, nil, fmt.Errorf("parse HTML template: %w", err)
	}
	cssTmpl, err := texttemplate.New(name+".css").Delims("@@", "@@").Parse(string(cssSource))
	if err != nil {
		return nil, nil, fmt.Errorf("parse CSS template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, page); err != nil {
		return nil, nil, fmt.Errorf("execute HTML template: %w", err)
	}
	var cssBuf bytes.Buffer
	if err := cssTmpl.Execute(&cssBuf, params.Theme.Colors); err != nil {
		return nil, nil, fmt.Errorf("execute CSS template: %w", err)
	}
	return htmlBuf.Bytes(), cssBuf.Bytes(), nil
}

// RenderHtml is kept for source compatibility with older callers.
//
//nolint:gocritic // Keep the original value-taking public API compatible.
func RenderHtml(params RenderingParams) ([]byte, []byte, error) {
	return RenderHTML(params)
}

//nolint:gocritic // HTML and CSS are a naturally paired return value.
func loadTemplatePair(root, category, name string) ([]byte, []byte, error) {
	htmlName := name + ".html"
	cssName := name + ".css"
	if root != "" {
		dir := filepath.Join(root, category)
		// category and name passed strict component validation before this call.
		htmlBytes, htmlErr := os.ReadFile(filepath.Join(dir, htmlName)) // #nosec G304
		cssBytes, cssErr := os.ReadFile(filepath.Join(dir, cssName))    // #nosec G304
		switch {
		case htmlErr == nil && cssErr == nil:
			return htmlBytes, cssBytes, nil
		case htmlErr == nil || cssErr == nil:
			return nil, nil, fmt.Errorf("custom template %s/%s must include both .html and .css files", category, name)
		case !errors.Is(htmlErr, os.ErrNotExist):
			return nil, nil, fmt.Errorf("read custom HTML template: %w", htmlErr)
		case !errors.Is(cssErr, os.ErrNotExist):
			return nil, nil, fmt.Errorf("read custom CSS template: %w", cssErr)
		}
	}

	if category != string(TemplateCategoryBase) || name != string(TemplateNameBase) {
		return nil, nil, fmt.Errorf("template %s/%s not found", category, name)
	}
	htmlBytes, err := fs.ReadFile(templateFS, filepath.ToSlash(filepath.Join("template", category, htmlName)))
	if err != nil {
		return nil, nil, fmt.Errorf("read built-in HTML template: %w", err)
	}
	cssBytes, err := fs.ReadFile(templateFS, filepath.ToSlash(filepath.Join("template", category, cssName)))
	if err != nil {
		return nil, nil, fmt.Errorf("read built-in CSS template: %w", err)
	}
	return htmlBytes, cssBytes, nil
}

func buildPageData(params *RenderingParams) PageData {
	active := -1
	for i, split := range params.Splits {
		if split.IsActive {
			active = i
			break
		}
	}

	start, end := 0, len(params.Splits)
	if active >= 0 {
		if limit := params.CounterSettings.DisplayPrev; limit.IsLimited {
			start = max(0, active-max(0, limit.Count))
		}
		if limit := params.CounterSettings.DisplayNext; limit.IsLimited {
			end = min(len(params.Splits), active+max(0, limit.Count)+1)
		}
	}

	data := PageData{RunName: params.Run.Name, CssName: params.CssFileName}
	data.Splits = make([]SplitData, 0, end-start)
	for _, split := range params.Splits[start:end] {
		data.Splits = append(data.Splits, SplitData{
			SplitName: split.Name,
			Hits:      split.Hits,
			PB:        split.PBHits,
			Diff:      split.Diff,
			IsActive:  split.IsActive,
		})
	}
	return data
}

// WriteOutput safely replaces the generated overlay files.
func WriteOutput(dir string, htmlBytes, cssBytes []byte) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create render directory: %w", err)
	}
	if err := writeAtomic(filepath.Join(dir, "output.html"), htmlBytes); err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(dir, "output.css"), cssBytes); err != nil {
		return err
	}
	return nil
}

func writeAtomic(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".ermokie-render-*")
	if err != nil {
		return fmt.Errorf("create temporary render file: %w", err)
	}
	tmp := file.Name()
	defer func() { _ = os.Remove(tmp) }()
	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return fmt.Errorf("set render permissions: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temporary render file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary render file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace %s: %w", filepath.Base(path), err)
	}
	return nil
}
