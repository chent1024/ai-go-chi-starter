package logging

import "time"

type Options struct {
	Level           string
	Format          string
	SourceEnabled   bool
	Output          string
	Dir             string
	RetentionDays   int
	CleanupInterval time.Duration
	Location        *time.Location
}

type OutboundOptions struct {
	Enabled bool
	Level   string
}

type OutboundEvent struct {
	Component string
	Target    string
	Method    string
	URL       string
	Status    int
	Latency   time.Duration
	BytesIn   int64
	BytesOut  int64
	Err       error
}
