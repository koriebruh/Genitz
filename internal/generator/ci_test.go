package generator

import (
	"strings"
	"testing"
)

func TestCIWorkflowContent(t *testing.T) {
	content := ciWorkflowContent("1.25")
	for _, want := range []string{`go-version: "1.25"`, "go build ./...", "go vet ./...", "go test ./...", "gofmt -l ."} {
		if !strings.Contains(content, want) {
			t.Errorf("ci workflow content missing %q\n---\n%s", want, content)
		}
	}
}
