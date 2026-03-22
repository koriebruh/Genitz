package generator

import "embed"

// templatesFS embeds all project scaffold templates so the binary is fully
// self-contained and can be run from any working directory after `go install`.
//
//go:embed templates
var templatesFS embed.FS
