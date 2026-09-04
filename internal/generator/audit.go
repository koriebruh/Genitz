package generator

import (
	"fmt"

	"github.com/koriebruh/Genitz/internal/tui"
)

// deprecatedReplacements maps a registry ID whose own maintainers have
// publicly stated it's in maintenance mode (not "considered legacy" by
// genitz's own opinion — hand-curated, cited, same "don't guess" discipline
// as license placeholders and the config-stub library list) to a suggested
// modern replacement ID plus why:
//   - gorilla/mux: the Gorilla Toolkit repo itself carries a maintenance-mode
//     notice recommending alternatives.
//   - logrus: its own README states it's in maintenance mode and recommends
//     zerolog/zap for new projects.
//
// Deliberately small — expand only with an equally citable source, not a
// personal opinion about what's "better."
var deprecatedReplacements = map[string]struct {
	ReplacementID string
	Why           string
}{
	"gorilla-mux": {ReplacementID: "chi", Why: "Gorilla Mux is in maintenance mode per its own repository notice"},
	"logrus":      {ReplacementID: "zap", Why: "logrus's own README states it is in maintenance mode and recommends a modern alternative"},
}

// AuditFinding is one issue AuditProject surfaced about an installed dependency.
type AuditFinding struct {
	ImportPath    string
	Name          string
	Kind          string // "deprecated" or "unmanaged"
	Detail        string
	ReplacementID string // set only for Kind == "deprecated"
}

// AuditProject cross-references dir's direct dependencies (via
// ListInstalled) against deprecatedReplacements, flagging any match — the
// reverse direction of the picker: instead of curating what to add, it
// curates what's already there. Also returns govulncheck's advisory, if
// available, so `genitz audit` is one place to check both curation quality
// and known vulnerabilities.
func AuditProject(dir string) (findings []AuditFinding, vulnAdvisory string, err error) {
	installed, err := ListInstalled(dir)
	if err != nil {
		return nil, "", err
	}

	for _, dep := range installed {
		if !dep.Managed {
			continue
		}
		id := ""
		for _, reg := range tui.DependencyRegistry {
			if reg.ImportPath == dep.ImportPath {
				id = reg.ID
				break
			}
		}
		if replacement, ok := deprecatedReplacements[id]; ok {
			findings = append(findings, AuditFinding{
				ImportPath:    dep.ImportPath,
				Name:          dep.Name,
				Kind:          "deprecated",
				Detail:        replacement.Why,
				ReplacementID: replacement.ReplacementID,
			})
		}
	}

	return findings, VulnCheckAdvisory(dir), nil
}

// PrintAudit renders AuditProject's results.
func PrintAudit(findings []AuditFinding, vulnAdvisory string) {
	if len(findings) == 0 {
		fmt.Println("✔ No deprecated dependencies found in the curated registry.")
	} else {
		fmt.Println("⚠ Stack health findings:")
		for _, f := range findings {
			replacement, ok := tui.FindByID(f.ReplacementID)
			suggestion := f.ReplacementID
			if ok {
				suggestion = replacement.Name
			}
			fmt.Printf("   %-24s %s — consider %s\n", f.Name, f.Detail, suggestion)
		}
	}

	if vulnAdvisory != "" {
		fmt.Println("\n" + vulnAdvisory)
	}
}
