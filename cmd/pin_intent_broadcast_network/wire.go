//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"pin_intent_broadcast_network/internal/biz"
	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/data"
	"pin_intent_broadcast_network/internal/p2p"
	"pin_intent_broadcast_network/internal/server"
	"pin_intent_broadcast_network/internal/service"
	"pin_intent_broadcast_network/internal/transport"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, p2p.ProviderSet, transport.ProviderSet, newApp))
}
