package mock_test

import (
	"github.com/TykTechnologies/benthos/v4/internal/bundle"
	"github.com/TykTechnologies/benthos/v4/internal/manager/mock"
)

var _ bundle.NewManagement = &mock.Manager{}
