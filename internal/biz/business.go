package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"pin_intent_broadcast_network/internal/biz/common"
)

// BusinessLogic provides unified access to all business logic components
type BusinessLogic struct {
	intentManager common.IntentManager
	validator     common.IntentValidator
	signer        common.IntentSigner
	processor     common.IntentProcessor
	matcher       common.IntentMatcher
	networkMgr    common.NetworkManager
	logger        *log.Helper
}

// NewBusinessLogic creates a new business logic instance
func NewBusinessLogic(
	intentManager common.IntentManager,
	validator common.IntentValidator,
	signer common.IntentSigner,
	processor common.IntentProcessor,
	matcher common.IntentMatcher,
	networkMgr common.NetworkManager,
	logger log.Logger,
) *BusinessLogic {
	return &BusinessLogic{
		intentManager: intentManager,
		validator:     validator,
		signer:        signer,
		processor:     processor,
		matcher:       matcher,
		networkMgr:    networkMgr,
		logger:        log.NewHelper(logger),
	}
}

// GetIntentManager returns the intent manager
func (bl *BusinessLogic) GetIntentManager() common.IntentManager {
	return bl.intentManager
}

// GetValidator returns the validator
func (bl *BusinessLogic) GetValidator() common.IntentValidator {
	return bl.validator
}

// GetSigner returns the signer
func (bl *BusinessLogic) GetSigner() common.IntentSigner {
	return bl.signer
}

// GetProcessor returns the processor
func (bl *BusinessLogic) GetProcessor() common.IntentProcessor {
	return bl.processor
}

// GetMatcher returns the matcher
func (bl *BusinessLogic) GetMatcher() common.IntentMatcher {
	return bl.matcher
}

// GetNetworkManager returns the network manager
func (bl *BusinessLogic) GetNetworkManager() common.NetworkManager {
	return bl.networkMgr
}

// Start starts all business logic components
func (bl *BusinessLogic) Start(ctx context.Context) error {
	bl.logger.Info("Starting business logic components")

	// Start components that need initialization
	if starter, ok := bl.intentManager.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return err
		}
	}

	if starter, ok := bl.processor.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return err
		}
	}

	if starter, ok := bl.networkMgr.(interface{ Start(context.Context) error }); ok {
		if err := starter.Start(ctx); err != nil {
			return err
		}
	}

	bl.logger.Info("Business logic components started successfully")
	return nil
}

// Stop stops all business logic components
func (bl *BusinessLogic) Stop(ctx context.Context) error {
	bl.logger.Info("Stopping business logic components")

	// Stop components that need cleanup
	if stopper, ok := bl.intentManager.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			bl.logger.Errorf("Failed to stop intent manager: %v", err)
		}
	}

	if stopper, ok := bl.processor.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			bl.logger.Errorf("Failed to stop processor: %v", err)
		}
	}

	if stopper, ok := bl.networkMgr.(interface{ Stop(context.Context) error }); ok {
		if err := stopper.Stop(ctx); err != nil {
			bl.logger.Errorf("Failed to stop network manager: %v", err)
		}
	}

	bl.logger.Info("Business logic components stopped")
	return nil
}

// GetHealthStatus returns the health status of all components
func (bl *BusinessLogic) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"status": "healthy",
	}

	// Collect health status from components
	if healthChecker, ok := bl.processor.(interface{ GetHealthStatus() map[string]interface{} }); ok {
		status["processor"] = healthChecker.GetHealthStatus()
	}

	if healthChecker, ok := bl.networkMgr.(interface{ GetMetrics() map[string]interface{} }); ok {
		status["network"] = healthChecker.GetMetrics()
	}

	return status
}
