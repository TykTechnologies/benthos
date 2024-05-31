package bloblang

import (
	"github.com/TykTechnologies/benthos/v4/internal/bloblang/plugins"
)

func init() {
	if err := plugins.Register(); err != nil {
		panic(err)
	}
}
