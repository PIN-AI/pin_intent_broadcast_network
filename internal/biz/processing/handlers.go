package processing

import (
	"fmt"
	"sync"

	"pin_intent_broadcast_network/internal/biz/common"
)

// HandlerRegistry implements the HandlerRegistry interface
// It manages the registration and retrieval of intent handlers
type HandlerRegistry struct {
	handlers map[string][]common.IntentHandler
	mu       sync.RWMutex
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string][]common.IntentHandler),
	}
}

// RegisterHandler registers an intent handler for a specific type
func (hr *HandlerRegistry) RegisterHandler(intentType string, handler common.IntentHandler) error {
	// TODO: Implement in task 5.3
	hr.mu.Lock()
	defer hr.mu.Unlock()

	handlers := hr.handlers[intentType]

	// Insert handler based on priority (higher priority first)
	inserted := false
	for i, h := range handlers {
		if handler.GetPriority() > h.GetPriority() {
			handlers = append(handlers[:i], append([]common.IntentHandler{handler}, handlers[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		handlers = append(handlers, handler)
	}

	hr.handlers[intentType] = handlers
	return nil
}

// UnregisterHandler unregisters an intent handler for a specific type
func (hr *HandlerRegistry) UnregisterHandler(intentType string) error {
	// TODO: Implement in task 5.3
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.handlers[intentType]; !exists {
		return fmt.Errorf("no handlers registered for intent type: %s", intentType)
	}

	delete(hr.handlers, intentType)
	return nil
}

// GetHandler retrieves the first handler for a specific intent type
func (hr *HandlerRegistry) GetHandler(intentType string) (common.IntentHandler, error) {
	// TODO: Implement in task 5.3
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	handlers, exists := hr.handlers[intentType]
	if !exists || len(handlers) == 0 {
		return nil, fmt.Errorf("no handler found for intent type: %s", intentType)
	}

	return handlers[0], nil
}

// ListHandlers returns all registered handlers
func (hr *HandlerRegistry) ListHandlers() map[string][]common.IntentHandler {
	// TODO: Implement in task 5.3
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	// Return a deep copy to prevent external modification
	result := make(map[string][]common.IntentHandler)
	for intentType, handlers := range hr.handlers {
		handlersCopy := make([]common.IntentHandler, len(handlers))
		copy(handlersCopy, handlers)
		result[intentType] = handlersCopy
	}

	return result
}

// GetHandlerCount returns the number of handlers for a specific type
func (hr *HandlerRegistry) GetHandlerCount(intentType string) int {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	handlers, exists := hr.handlers[intentType]
	if !exists {
		return 0
	}

	return len(handlers)
}

// GetAllHandlerTypes returns all registered intent types
func (hr *HandlerRegistry) GetAllHandlerTypes() []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	types := make([]string, 0, len(hr.handlers))
	for intentType := range hr.handlers {
		types = append(types, intentType)
	}

	return types
}
