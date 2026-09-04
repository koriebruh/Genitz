package generator

import (
	"fmt"
	"sort"
	"strings"
)

// readmeContent returns a minimal README.md for req: project name, module
// path, a quickstart, and the chosen dependency list — enough to orient
// someone opening the repo cold, without guessing at a fuller project
// description genitz has no way to know.
func readmeContent(req Requirement) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", req.ProjectName)
	fmt.Fprintf(&b, "Scaffolded with [Genitz](https://github.com/koriebruh/Genitz).\n\n")
	b.WriteString("## Getting Started\n\n")
	b.WriteString("```sh\ngo run .\n```\n\n")
	fmt.Fprintf(&b, "Module: `%s`\n", req.PackageName)

	if len(req.Deps) > 0 {
		names := make([]string, 0, len(req.Deps))
		for _, dep := range req.Deps {
			names = append(names, dep.Name)
		}
		sort.Strings(names)

		b.WriteString("\n## Dependencies\n\n")
		for _, name := range names {
			fmt.Fprintf(&b, "- %s\n", name)
		}
	}

	return b.String()
}
