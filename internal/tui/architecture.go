package tui

const (
	ArchStandard = "Standard Layout"
	ArchMicro    = "Microservice"
	ArchClean    = "Clean Architecture"
	ArchDDD      = "Domain Driven Design"
	ArchCLI      = "CLI Tool"
)

type Architecture struct {
	Name        string
	Description string
	TemplateDir string
}

type Step int

const (
	StepSplash Step = iota
	StepFolder
	StepPackage
	StepArch
	StepDeps
)
