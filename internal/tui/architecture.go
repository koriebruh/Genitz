package tui

const (
	ArchStandard = "Standard Layout"
	ArchMicro    = "Microservice"
	ArchClean    = "Clean Architecture"
	ArchDDD      = "Domain Driven Design"
	ArchCLI      = "CLI Tool"
)

// Architecture holds metadata for a project structure template.
type Architecture struct {
	Name        string
	Description string
	TemplateDir string
}

// Step represents a wizard step in the TUI flow.
type Step int

const (
	StepFolder Step = iota // removed StepSplash — logo is now a persistent header
	StepPackage
	StepArch
	StepDeps
	StepReview
	StepGenerating
)

// AvailableArchitectures is the set of architecture names that have a complete
// template. Unlisted names will be shown as "coming soon" in the TUI.
var AvailableArchitectures = map[string]bool{
	ArchMicro:    true,
	ArchClean:    true,
	ArchStandard: true,
	ArchDDD:      true,
	ArchCLI:      true,
}

// archDescriptions maps each architecture name to a short description
// shown below the option in the architecture selection panel.
var archDescriptions = map[string]string{
	ArchStandard: "cmd/ · internal/ · pkg/  — idiomatic Go layout",
	ArchMicro:    "Layered service with clear domain boundaries",
	ArchClean:    "entity/ · usecase/ · repository/ · delivery/",
	ArchDDD:      "domain/ · application/ · infra/ — bounded contexts",
	ArchCLI:      "Single main package, ideal for small CLIs",
}

// archTrees maps each architecture name to a visual ASCII folder tree
// shown when the item is highlighted.
var archTrees = map[string]string{
	ArchStandard: "├── cmd/\n├── internal/\n└── pkg/",
	ArchMicro:    "├── cmd/\n├── config/\n├── internal/\n│   ├── api/\n│   └── core/\n└── pkg/",
	ArchClean:    "├── domain/\n├── usecase/\n├── repository/\n└── delivery/",
	ArchDDD:      "├── domain/\n├── application/\n└── infrastructure/",
	ArchCLI:      "├── cmd/\n└── pkg/",
}
