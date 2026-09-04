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
		_, ok := licenseContent(kind)
		wantContent := kind == "mit" || kind == "apache-2.0"
		if ok != wantContent {
			t.Errorf("licenseContent(%q) ok=%v, want %v", kind, ok, wantContent)
		}
	}
}
