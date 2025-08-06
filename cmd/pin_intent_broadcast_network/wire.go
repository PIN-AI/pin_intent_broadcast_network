//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"pin_intent_broadcast_network/internal/biz"
	"pin_intent_broadcast_network/internal/biz/execution"
	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/data"
	"pin_intent_broadcast_network/internal/p2p"
	"pin_intent_broadcast_network/internal/server"
	"pin_intent_broadcast_network/internal/service"
	"pin_intent_broadcast_network/internal/transport"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// wireApp init kratos application.
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		// Extract individual configs from Bootstrap
		wire.FieldsOf(new(*conf.Bootstrap), "Server", "Data"),

		// Core providers
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,

		// P2P and Transport providers (they use full bootstrap config)
		p2p.ProviderSet,
		transport.ProviderSet,

		// Execution automation providers
		execution.ProviderSet,

		// Zap logger provider
		NewZapLogger,

		newApp,
	))
}

// NewZapLogger creates a zap logger from kratos logger
func NewZapLogger(logger log.Logger) *zap.Logger {
	// For now, create a simple zap logger
	// In production, you might want to configure this more carefully
	zapLogger, _ := zap.NewProduction()
	return zapLogger
}
