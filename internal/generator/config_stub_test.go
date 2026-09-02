package generator

import (
	"strings"
	"testing"

	"github.com/koriebruh/Genitz/internal/tui"
)

func TestDetectConfigLibSingleMatch(t *testing.T) {
	deps := map[int]tui.Dependency{
		0: {ID: "viper"},
		1: {ID: "fiber"},
	}
	dep, ok := detectConfigLib(deps)
	if !ok || dep.ID != "viper" {
		t.Fatalf("got %+v, ok=%v; want viper, ok=true", dep, ok)
	}
}

func TestDetectConfigLibNoMatch(t *testing.T) {
	deps := map[int]tui.Dependency{0: {ID: "fiber"}}
	if _, ok := detectConfigLib(deps); ok {
		t.Fatal("expected no match")
	}
}

func TestDetectConfigLibAmbiguous(t *testing.T) {
	deps := map[int]tui.Dependency{
		0: {ID: "viper"},
		1: {ID: "godotenv"},
	}
	if _, ok := detectConfigLib(deps); ok {
		t.Fatal("expected ambiguous (2 config libs) to skip generation")
	}
}

func TestConfigStubContentAllFourLibs(t *testing.T) {
	cases := []struct {
		id      string
		imports string
	}{
		{"viper", "github.com/spf13/viper"},
		{"godotenv", "github.com/joho/godotenv"},
		{"envconfig", "github.com/kelseyhightower/envconfig"},
		{"koanf", "github.com/knadh/koanf/v2"},
	}
	for _, c := range cases {
		content, ok := configStubContent(tui.Dependency{ID: c.id})
		if !ok {
			t.Errorf("%s: expected ok=true", c.id)
			continue
		}
		if !strings.Contains(content, "package config") {
			t.Errorf("%s: missing package declaration", c.id)
		}
		if !strings.Contains(content, c.imports) {
			t.Errorf("%s: missing import %q\n---\n%s", c.id, c.imports, content)
		}
	}
}

func TestConfigStubContentUnknownLib(t *testing.T) {
	if _, ok := configStubContent(tui.Dependency{ID: "fiber"}); ok {
		t.Fatal("expected ok=false for a non-config dependency")
	}
}
