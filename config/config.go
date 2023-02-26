package config

import (
	"context"
)

// Config -
type Config struct {
	Project       string
	Credentials   string
	Zones         []string
	Regions       []string
	Timeout       int
	PollTime      int
	Context       context.Context
	NoDryRun      bool
	NoKeepProject bool
}
