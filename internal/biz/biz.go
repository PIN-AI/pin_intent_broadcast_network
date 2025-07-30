package biz

import (
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/intent"
	"pin_intent_broadcast_network/internal/biz/matching"
	"pin_intent_broadcast_network/internal/biz/network"
	"pin_intent_broadcast_network/internal/biz/processing"
	"pin_intent_broadcast_network/internal/biz/security"
	"pin_intent_broadcast_network/internal/biz/validation"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGreeterUsecase,
	NewBusinessLogic,
	intent.NewManager,
	validation.NewValidator,
	security.NewSigner,
	processing.NewProcessor,
	matching.NewMatcher,
	network.NewManager,
)
