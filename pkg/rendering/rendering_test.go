package rendering

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	overlaythemes "codeberg.org/veya/ermokie/pkg/rendering/overlayThemes"
	"codeberg.org/veya/ermokie/pkg/types"
)

func testParams() RenderingParams {
	return RenderingParams{
		Run: types.Run{Name: "Any <script>bad()</script>"},
		Splits: []types.Split{
			{Name: "one", Hits: 1, PBHits: 2},
			{Name: "two", Hits: 3, PBHits: 3, IsActive: true},
			{Name: "three", Hits: 4, PBHits: 5},
			{Name: "four", Hits: 6, PBHits: 8},
		},
		Tc:          TemplateCategoryBase,
		Tn:          TemplateNameBase,
		CssFileName: "output.css",
		Theme:       overlaythemes.NewOverlayThemes().GetThemeByName("default"),
	}
}

func TestRenderHTMLUsesWindowAroundActiveSplitAndEscapesData(t *testing.T) {
	params := testParams()
	params.CounterSettings.DisplayPrev = DisplayCountSettingsData{IsLimited: true, Count: 1}
	params.CounterSettings.DisplayNext = DisplayCountSettingsData{IsLimited: true, Count: 1}

	html, css, err := RenderHTML(params)
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	body := string(html)
	if strings.Contains(body, ">four<") {
		t.Fatalf("rendered rows outside the requested window: %s", body)
	}
	for _, want := range []string{">one<", ">two<", ">three<", "Any &lt;script&gt;bad()&lt;/script&gt;"} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered HTML missing %q", want)
		}
	}
	if !strings.Contains(string(css), "--color-bg-highlighted") {
		t.Fatal("rendered CSS did not apply the theme")
	}
}

func TestRenderHTMLLoadsCustomTemplatePair(t *testing.T) {
	dir := t.TempDir()
	customDir := filepath.Join(dir, "compact")
	if err := os.MkdirAll(customDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customDir, "stream.html"), []byte(`{{.RunName}}|{{range .Splits}}{{.SplitName}},{{end}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customDir, "stream.css"), []byte(`color: @@.TextPrimary@@;`), 0o600); err != nil {
		t.Fatal(err)
	}

	params := testParams()
	params.TemplateDir = dir
	params.Tc = "compact"
	params.Tn = "stream"
	html, css, err := RenderHTML(params)
	if err != nil {
		t.Fatalf("RenderHTML(custom) error = %v", err)
	}
	if !strings.Contains(string(html), "Any &lt;script&gt;bad()") || !strings.Contains(string(html), "one,two,three,four,") {
		t.Fatalf("unexpected custom HTML: %s", html)
	}
	if got := string(css); got != "color: #222222;" {
		t.Fatalf("custom CSS = %q", got)
	}
}

func TestRenderHTMLRejectsTemplateTraversal(t *testing.T) {
	params := testParams()
	params.TemplateDir = t.TempDir()
	params.Tc = ".."
	if _, _, err := RenderHTML(params); err == nil {
		t.Fatal("template traversal was accepted")
	}
}

func TestWriteOutputAtomicallyCreatesPair(t *testing.T) {
	dir := t.TempDir()
	if err := WriteOutput(dir, []byte("html"), []byte("css")); err != nil {
		t.Fatalf("WriteOutput() error = %v", err)
	}
	for name, want := range map[string]string{"output.html": "html", "output.css": "css"} {
		got, err := os.ReadFile(filepath.Join(dir, name)) // #nosec G304 -- names are fixed test fixtures.
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q", name, got)
		}
	}
}
