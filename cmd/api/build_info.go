package main

import "ai-go-chi-starter/internal/runtime"

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func buildInfo() runtime.BuildInfo {
	return runtime.BuildInfo{
		Service:   "api",
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
}
