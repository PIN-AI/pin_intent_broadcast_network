package execution

import (
	"github.com/google/wire"
)

// ProviderSet is the wire provider set for execution module
var ProviderSet = wire.NewSet(
	NewAutomationManager,
)