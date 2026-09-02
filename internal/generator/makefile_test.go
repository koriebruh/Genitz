package generator

import (
	"strings"
	"testing"
)

func TestMakefileContentPlain(t *testing.T) {
	content := makefileContent(false, false)
	for _, want := range []string{"build:", "test:", "run:", "fmt:", "vet:"} {
		if !strings.Contains(content, want) {
			t.Errorf("makefile missing %q\n---\n%s", want, content)
		}
	}
	if strings.Contains(content, "docker") {
		t.Error("plain makefile should not mention docker")
	}
}

func TestMakefileContentDockerfileOnly(t *testing.T) {
	content := makefileContent(true, false)
	if !strings.Contains(content, "docker-build:") || !strings.Contains(content, "docker-run:") {
		t.Errorf("expected docker-build/docker-run targets\n---\n%s", content)
	}
	if strings.Contains(content, "docker-up:") {
		t.Error("should not have docker-up without a compose file")
	}
}

func TestMakefileContentWithCompose(t *testing.T) {
	content := makefileContent(true, true)
	if !strings.Contains(content, "docker-up:") || !strings.Contains(content, "docker-down:") {
		t.Errorf("expected docker-up/docker-down targets\n---\n%s", content)
	}
	if strings.Contains(content, "docker-build:") {
		t.Error("should prefer compose targets over plain docker-build when compose exists")
	}
}
