package types

import (
	"time"
)

// Timeouts is a global structure that contains configuration for zipper Timeouts
type Timeouts struct {
	Find    time.Duration `yaml:"find"`
	Expand  time.Duration `yaml:"expand"`
	Render  time.Duration `yaml:"render"`
	Connect time.Duration `yaml:"connect"`
}
