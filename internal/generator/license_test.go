package generator

import "testing"

func TestValidLicenseKind(t *testing.T) {
	cases := map[string]bool{
		"":           true,
		"mit":        true,
		"apache-2.0": true,
		"Mit":        false,
		"apache":     false,
		"bsd":        false,
	}
	for kind, want := range cases {
		if got := ValidLicenseKind(kind); got != want {
			t.Errorf("ValidLicenseKind(%q) = %v, want %v", kind, got, want)
		}
	}
}

func TestLicenseContentMatchesValidLicenseKind(t *testing.T) {
	for kind := range validLicenseKinds {
		_, ok := licenseContent(kind, "")
		wantContent := kind == "mit" || kind == "apache-2.0"
		if ok != wantContent {
			t.Errorf("licenseContent(%q) ok=%v, want %v", kind, ok, wantContent)
		}
	}
}

func TestLicenseContentFillsHolderWhenProvided(t *testing.T) {
	content, ok := licenseContent("mit", "Jane Doe")
	if !ok {
		t.Fatal("expected licenseContent(\"mit\", ...) to succeed")
	}
	if !contains(content, "Jane Doe") {
		t.Fatalf("expected the holder to appear in the license body, got:\n%s", content)
	}
	if contains(content, "[COPYRIGHT HOLDER]") {
		t.Fatal("expected the placeholder to be replaced when a holder is provided")
	}
}

func TestLicenseContentLeavesPlaceholderWhenHolderEmpty(t *testing.T) {
	content, ok := licenseContent("apache-2.0", "")
	if !ok {
		t.Fatal("expected licenseContent(\"apache-2.0\", ...) to succeed")
	}
	if !contains(content, "[COPYRIGHT HOLDER]") {
		t.Fatalf("expected the placeholder to remain when no holder is provided, got:\n%s", content)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
