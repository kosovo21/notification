package version

// These variables are set at compile time via -ldflags.
//
//	go build -ldflags "-X notification-system/internal/version.Version=v1.0.0 \
//	                    -X notification-system/internal/version.Commit=abc1234 \
//	                    -X notification-system/internal/version.BuildDate=2026-02-16T06:00:00Z"
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info returns version information as a map, suitable for JSON responses.
func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     Commit,
		"build_date": BuildDate,
	}
}
