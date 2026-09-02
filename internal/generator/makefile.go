package generator

import "strings"

// makefileContent returns a Makefile with the standard build/test/run/fmt/vet
// targets, plus docker targets that match whatever Docker actually produced:
// docker-up/docker-down when a compose file exists, otherwise
// docker-build/docker-run when just a bare Dockerfile does.
func makefileContent(hasDockerfile, hasCompose bool) string {
	var b strings.Builder

	b.WriteString(".PHONY: build test run fmt vet")
	switch {
	case hasCompose:
		b.WriteString(" docker-up docker-down")
	case hasDockerfile:
		b.WriteString(" docker-build docker-run")
	}
	b.WriteString("\n\n")

	b.WriteString("build:\n\tgo build -o bin/app .\n\n")
	b.WriteString("test:\n\tgo test ./...\n\n")
	b.WriteString("run:\n\tgo run .\n\n")
	b.WriteString("fmt:\n\tgofmt -w .\n\n")
	b.WriteString("vet:\n\tgo vet ./...\n")

	switch {
	case hasCompose:
		b.WriteString("\ndocker-up:\n\tdocker compose up --build\n\n")
		b.WriteString("docker-down:\n\tdocker compose down\n")
	case hasDockerfile:
		b.WriteString("\ndocker-build:\n\tdocker build -t app .\n\n")
		b.WriteString("docker-run:\n\tdocker run --rm -p 8080:8080 app\n")
	}

	return b.String()
}
