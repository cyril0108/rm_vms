package buildinfo

// Version Info with default values
// Should be set by GO_LDFLAGS
var (
    Version   = "dev"
    CommitSHA = "none"
    BuildTime = "unknown"
)