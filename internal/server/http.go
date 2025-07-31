package server

import (
	"encoding/json"
	nethttp "net/http"

	v1 "pin_intent_broadcast_network/api/helloworld/v1"
	intentv1 "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, intent *service.IntentService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	intentv1.RegisterIntentServiceHTTPServer(srv, intent)

	// Add debug endpoints for intent monitoring
	addDebugEndpoints(srv, logger)

	return srv
}

// addDebugEndpoints adds debug endpoints for monitoring
func addDebugEndpoints(srv *http.Server, logger log.Logger) {
	helper := log.NewHelper(logger)

	// Intent monitoring config endpoint
	srv.HandleFunc("/debug/intent-monitoring/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return mock config for now - in a real implementation this would come from the actual config
		config := map[string]interface{}{
			"subscription_mode":  "all",
			"statistics_enabled": true,
			"wildcard_patterns":  []string{"intent-broadcast.*"},
			"explicit_topics":    []string{},
		}

		json.NewEncoder(w).Encode(config)
		helper.Debug("Served intent monitoring config")
	})

	// Intent monitoring subscriptions endpoint
	srv.HandleFunc("/debug/intent-monitoring/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return mock subscription status - in a real implementation this would come from the subscription manager
		status := map[string]interface{}{
			"mode": "all",
			"active_subscriptions": []string{
				"intent-broadcast.trade",
				"intent-broadcast.swap",
				"intent-broadcast.exchange",
				"intent-broadcast.transfer",
				"intent-broadcast.send",
				"intent-broadcast.payment",
				"intent-broadcast.lending",
				"intent-broadcast.borrow",
				"intent-broadcast.loan",
				"intent-broadcast.investment",
				"intent-broadcast.staking",
				"intent-broadcast.yield",
				"intent-broadcast.general",
				"intent-broadcast.matching",
				"intent-broadcast.notification",
				"intent-broadcast.status",
			},
			"total_messages": 0,
			"total_errors":   0,
		}

		json.NewEncoder(w).Encode(status)
		helper.Debug("Served intent monitoring subscriptions")
	})

	// Intent monitoring stats endpoint
	srv.HandleFunc("/debug/intent-monitoring/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return mock statistics - in a real implementation this would come from the statistics manager
		stats := map[string]interface{}{
			"total_received":   0,
			"total_filtered":   0,
			"total_duplicates": 0,
			"by_type":          map[string]interface{}{},
			"by_sender":        map[string]interface{}{},
			"by_topic":         map[string]interface{}{},
		}

		json.NewEncoder(w).Encode(stats)
		helper.Debug("Served intent monitoring stats")
	})

	// P2P status endpoint
	srv.HandleFunc("/debug/p2p/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return mock P2P status
		status := map[string]interface{}{
			"peer_id":          "12D3KooWExample...",
			"connected_peers":  2,
			"listen_addresses": []string{"/ip4/0.0.0.0/tcp/9001"},
		}

		json.NewEncoder(w).Encode(status)
		helper.Debug("Served P2P status")
	})

	// Topics subscriptions endpoint
	srv.HandleFunc("/debug/topics/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return mock topic subscriptions
		subscriptions := map[string]interface{}{
			"subscribed_topics": []string{
				"intent-broadcast.trade",
				"intent-broadcast.swap",
				"intent-broadcast.exchange",
				"intent-broadcast.transfer",
				"intent-broadcast.send",
				"intent-broadcast.payment",
				"intent-broadcast.lending",
				"intent-broadcast.borrow",
				"intent-broadcast.loan",
				"intent-broadcast.investment",
				"intent-broadcast.staking",
				"intent-broadcast.yield",
				"intent-broadcast.general",
				"intent-broadcast.matching",
				"intent-broadcast.notification",
				"intent-broadcast.status",
			},
		}

		json.NewEncoder(w).Encode(subscriptions)
		helper.Debug("Served topics subscriptions")
	})

	// P2P peers endpoint
	srv.HandleFunc("/debug/p2p/peers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return mock peer connections
		peers := map[string]interface{}{
			"peers": []map[string]interface{}{
				{
					"peer_id": "12D3KooWPeer1...",
					"address": "/ip4/127.0.0.1/tcp/9001",
				},
				{
					"peer_id": "12D3KooWPeer2...",
					"address": "/ip4/127.0.0.1/tcp/9002",
				},
			},
		}

		json.NewEncoder(w).Encode(peers)
		helper.Debug("Served P2P peers")
	})

	// Health check endpoint
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}
