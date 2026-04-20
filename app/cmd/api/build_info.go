package main

import apimetrics "ai-go-chi-starter/internal/transport/httpapi/metrics"

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func buildInfo() apimetrics.BuildInfo {
	return apimetrics.BuildInfo{
		Service:   "api",
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
}
