package generator

// Version is Genitz's own release version. The release pipeline
// (.goreleaser.yaml) overrides it at build time via
// -ldflags "-X .../generator.Version=vX.Y.Z" — ldflags -X can only target a
// var, not a const, hence this being one. `go install`/local builds that
// skip that step just get this literal default.
var Version = "0.1.0-dev"
