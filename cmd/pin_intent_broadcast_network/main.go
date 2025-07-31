package main

import (
	"context"
	"flag"
	"os"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/biz/intent"
	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/p2p"
	"pin_intent_broadcast_network/internal/transport"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, networkManager p2p.NetworkManager, transportManager transport.TransportManager, intentManager common.IntentManager, bootstrap *conf.Bootstrap) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		// Add P2P network lifecycle hooks
		kratos.BeforeStart(func(ctx context.Context) error {
			// Set global intent manager reference
			if intentManager != nil {
				intent.SetIntentManager(intentManager)
				logger.Log(log.LevelInfo, "msg", "Intent manager reference set")
			}

			// Start P2P network manager
			if networkManager != nil {
				// Get P2P config from bootstrap configuration
				p2pConfig, err := p2p.NewP2PConfig(bootstrap)
				if err != nil {
					logger.Log(log.LevelError, "msg", "Failed to create P2P config", "error", err)
					return err
				}

				if err := networkManager.Start(ctx, p2pConfig); err != nil {
					logger.Log(log.LevelError, "msg", "Failed to start P2P network", "error", err)
					return err
				}
				logger.Log(log.LevelInfo, "msg", "P2P network started successfully")
			}

			// Create and start transport manager with the started P2P host
			if hostManager := networkManager.GetHostManager(); hostManager != nil {
				if host := hostManager.GetHost(); host != nil {
					// Create a new transport manager with the actual host
					zapLogger := NewZapLogger(logger)
					actualTransportManager := transport.NewTransportManager(host, zapLogger)

					transportConfig := &transport.TransportConfig{
						EnableGossipSub:                   true,
						GossipSubHeartbeatInterval:        time.Second,
						GossipSubD:                        6,
						GossipSubDLo:                      4,
						GossipSubDHi:                      12,
						GossipSubFanoutTTL:                time.Minute,
						EnableMessageSigning:              true,
						EnableStrictSignatureVerification: true,
						MessageIDCacheSize:                1000,
						MessageTTL:                        5 * time.Minute,
						MaxMessageSize:                    1048576,
					}

					if err := actualTransportManager.Start(ctx, transportConfig); err != nil {
						logger.Log(log.LevelError, "msg", "Failed to start transport manager", "error", err)
						return err
					}
					logger.Log(log.LevelInfo, "msg", "Transport manager started successfully")

					// Replace the nil transport manager with the actual one
					transportManager = actualTransportManager

					// Update the intent manager with the actual transport manager
					if intentManager != nil {
						if manager, ok := intentManager.(*intent.Manager); ok {
							manager.SetTransportManager(actualTransportManager)
							logger.Log(log.LevelInfo, "msg", "Intent manager transport updated")

							// Start intent subscription
							if err := manager.StartIntentSubscription(ctx); err != nil {
								logger.Log(log.LevelError, "msg", "Failed to start intent subscription", "error", err)
							} else {
								logger.Log(log.LevelInfo, "msg", "Intent subscription started")
							}
						}
					}
				}
			}

			return nil
		}),
		kratos.AfterStop(func(ctx context.Context) error {
			// Stop transport manager
			if transportManager != nil {
				if err := transportManager.Stop(); err != nil {
					logger.Log(log.LevelError, "msg", "Failed to stop transport manager", "error", err)
				}
			}

			// Stop P2P network manager
			if networkManager != nil {
				if err := networkManager.Stop(); err != nil {
					logger.Log(log.LevelError, "msg", "Failed to stop P2P network", "error", err)
				}
			}

			return nil
		}),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(&bc, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
