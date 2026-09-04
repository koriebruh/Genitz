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

	if diagram, ok := architectureDiagram(req); ok {
		b.WriteString("\n## Architecture\n\n")
		b.WriteString(diagram)
	}

	return b.String()
}

// architectureDiagram returns a Mermaid flowchart of req's selected
// dependencies grouped by category — derived entirely from data already
// selected (dep.Category, same field the picker/review screen group by),
// no new curated data needed. ok is false with zero deps, same
// nothing-to-show convention as composeContent/envExampleContent.
func architectureDiagram(req Requirement) (string, bool) {
	if len(req.Deps) == 0 {
		return "", false
	}

	byCategory := make(map[string][]struct{ id, name string })
	var categories []string
	for _, dep := range req.Deps {
		if _, seen := byCategory[dep.Category]; !seen {
			categories = append(categories, dep.Category)
		}
		byCategory[dep.Category] = append(byCategory[dep.Category], struct{ id, name string }{dep.ID, dep.Name})
	}
	sort.Strings(categories)

	appNode := mermaidNodeID(req.ProjectName)
	if appNode == "" {
		appNode = "App"
	}

	var b strings.Builder
	b.WriteString("```mermaid\nflowchart LR\n")
	fmt.Fprintf(&b, "    %s[%s]\n", appNode, req.ProjectName)
	for _, cat := range categories {
		deps := byCategory[cat]
		sort.Slice(deps, func(i, j int) bool { return deps[i].name < deps[j].name })
		for _, dep := range deps {
			node := mermaidNodeID(dep.id)
			fmt.Fprintf(&b, "    %s --> %s[%s]\n", appNode, node, dep.name)
		}
	}
	b.WriteString("```\n")

	return b.String(), true
}

// mermaidNodeID sanitizes s into a safe Mermaid flowchart node identifier —
// hyphens (common in registry IDs like "postgres-driver") aren't reliably
// safe across Mermaid versions, so they're replaced with underscores.
func mermaidNodeID(s string) string {
	return strings.NewReplacer("-", "_", " ", "_", ".", "_", "/", "_").Replace(s)
}
