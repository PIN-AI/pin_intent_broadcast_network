package transport

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// MessageRouter handles message routing and deduplication
type MessageRouter interface {
	// RouteMessage routes message based on routing rules
	RouteMessage(ctx context.Context, topic string, msg *TransportMessage) error
	// DeduplicateMessage checks if message is duplicate
	DeduplicateMessage(msg *TransportMessage) bool
	// AddFilter adds message filter
	AddFilter(filter MessageFilter)
	// RemoveFilter removes message filter
	RemoveFilter(filterID string)
	// GetRouteCount returns message route count
	GetRouteCount() int64
	// GetDuplicateCount returns duplicate message count
	GetDuplicateCount() int64
	// Start starts the message router
	Start(ctx context.Context) error
	// Stop stops the message router
	Stop() error
}

// MessageFilter filters messages
type MessageFilter interface {
	// ID returns filter ID
	ID() string
	// Filter filters message, returns true to pass, false to drop
	Filter(msg *TransportMessage) bool
	// Priority returns filter priority (higher number = higher priority)
	Priority() int
}

// messageRouter message router implementation
type messageRouter struct {
	// Message deduplication
	messageCache    map[string]*cacheEntry
	cacheMu         sync.RWMutex
	cacheSize       int
	cacheTTL        time.Duration
	
	// Message filters
	filters         []MessageFilter
	filtersMu       sync.RWMutex
	
	// Routing statistics
	routeCount      int64
	duplicateCount  int64
	droppedCount    int64
	
	// Router state
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   bool
	logger      *zap.Logger
}

// cacheEntry represents a cached message entry
type cacheEntry struct {
	messageID string
	timestamp time.Time
	sender    peer.ID
	topic     string
}

// messageFilter basic message filter implementation
type messageFilter struct {
	id       string
	priority int
	filterFn func(*TransportMessage) bool
}

// NewMessageRouter creates new message router
func NewMessageRouter(cacheSize int, cacheTTL time.Duration, logger *zap.Logger) MessageRouter {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	if cacheSize <= 0 {
		cacheSize = 10000 // Default cache size
	}
	
	if cacheTTL <= 0 {
		cacheTTL = 10 * time.Minute // Default cache TTL
	}
	
	return &messageRouter{
		messageCache: make(map[string]*cacheEntry),
		cacheSize:    cacheSize,
		cacheTTL:     cacheTTL,
		filters:      make([]MessageFilter, 0),
		logger:       logger.Named("message_router"),
	}
}

// Start starts the message router
func (mr *messageRouter) Start(ctx context.Context) error {
	if mr.isRunning {
		return fmt.Errorf("message router already running")
	}
	
	mr.ctx, mr.cancel = context.WithCancel(ctx)
	mr.isRunning = true
	
	// Start cache cleanup goroutine
	go mr.cacheCleaner()
	
	mr.logger.Info("Message router started",
		zap.Int("cache_size", mr.cacheSize),
		zap.Duration("cache_ttl", mr.cacheTTL),
	)
	
	return nil
}

// Stop stops the message router
func (mr *messageRouter) Stop() error {
	if !mr.isRunning {
		return fmt.Errorf("message router not running")
	}
	
	if mr.cancel != nil {
		mr.cancel()
	}
	
	mr.isRunning = false
	mr.logger.Info("Message router stopped")
	
	return nil
}

// RouteMessage routes message based on routing rules
func (mr *messageRouter) RouteMessage(ctx context.Context, topic string, msg *TransportMessage) error {
	if !mr.isRunning {
		return fmt.Errorf("message router not running")
	}
	
	// Check for duplicates first
	if mr.DeduplicateMessage(msg) {
		mr.duplicateCount++
		mr.logger.Debug("Duplicate message detected, dropping",
			zap.String("message_id", msg.ID),
			zap.String("topic", FormatTopic(topic)),
		)
		return &TransportError{
			Code:    "DUPLICATE_MESSAGE",
			Message: "Message is duplicate and will be dropped",
			Details: fmt.Sprintf("message_id: %s, topic: %s", msg.ID, topic),
		}
	}
	
	// Apply message filters
	if !mr.applyFilters(msg) {
		mr.droppedCount++
		mr.logger.Debug("Message dropped by filters",
			zap.String("message_id", msg.ID),
			zap.String("topic", FormatTopic(topic)),
		)
		return &TransportError{
			Code:    "MESSAGE_FILTERED",
			Message: "Message dropped by filter rules",
			Details: fmt.Sprintf("message_id: %s, topic: %s", msg.ID, topic),
		}
	}
	
	// Add to cache
	mr.addToCache(msg, topic)
	mr.routeCount++
	
	mr.logger.Debug("Message routed successfully",
		zap.String("message_id", msg.ID),
		zap.String("topic", FormatTopic(topic)),
		zap.String("sender", msg.Sender),
	)
	
	return nil
}

// DeduplicateMessage checks if message is duplicate
func (mr *messageRouter) DeduplicateMessage(msg *TransportMessage) bool {
	if msg == nil || msg.ID == "" {
		return false
	}
	
	mr.cacheMu.RLock()
	_, exists := mr.messageCache[msg.ID]
	mr.cacheMu.RUnlock()
	
	return exists
}

// AddFilter adds message filter
func (mr *messageRouter) AddFilter(filter MessageFilter) {
	if filter == nil {
		return
	}
	
	mr.filtersMu.Lock()
	defer mr.filtersMu.Unlock()
	
	// Remove existing filter with same ID
	mr.removeFilterByID(filter.ID())
	
	// Add new filter
	mr.filters = append(mr.filters, filter)
	
	// Sort filters by priority (highest first)
	mr.sortFilters()
	
	mr.logger.Info("Message filter added",
		zap.String("filter_id", filter.ID()),
		zap.Int("priority", filter.Priority()),
		zap.Int("total_filters", len(mr.filters)),
	)
}

// RemoveFilter removes message filter
func (mr *messageRouter) RemoveFilter(filterID string) {
	mr.filtersMu.Lock()
	defer mr.filtersMu.Unlock()
	
	removed := mr.removeFilterByID(filterID)
	if removed {
		mr.logger.Info("Message filter removed",
			zap.String("filter_id", filterID),
			zap.Int("remaining_filters", len(mr.filters)),
		)
	}
}

// GetRouteCount returns message route count
func (mr *messageRouter) GetRouteCount() int64 {
	return mr.routeCount
}

// GetDuplicateCount returns duplicate message count
func (mr *messageRouter) GetDuplicateCount() int64 {
	return mr.duplicateCount
}

// GetDroppedCount returns dropped message count
func (mr *messageRouter) GetDroppedCount() int64 {
	return mr.droppedCount
}

// addToCache adds message to deduplication cache
func (mr *messageRouter) addToCache(msg *TransportMessage, topic string) {
	mr.cacheMu.Lock()
	defer mr.cacheMu.Unlock()
	
	// Check cache size limit
	if len(mr.messageCache) >= mr.cacheSize {
		// Remove oldest entries (simple LRU-like cleanup)
		mr.cleanupOldEntries(mr.cacheSize / 4) // Remove 25% of entries
	}
	
	// Parse sender as peer ID
	senderID, err := peer.Decode(msg.Sender)
	if err != nil {
		mr.logger.Debug("Failed to decode sender peer ID",
			zap.String("sender", msg.Sender),
			zap.Error(err),
		)
		senderID = ""
	}
	
	// Add to cache
	mr.messageCache[msg.ID] = &cacheEntry{
		messageID: msg.ID,
		timestamp: time.Now(),
		sender:    senderID,
		topic:     topic,
	}
}

// applyFilters applies all filters to message
func (mr *messageRouter) applyFilters(msg *TransportMessage) bool {
	mr.filtersMu.RLock()
	filters := make([]MessageFilter, len(mr.filters))
	copy(filters, mr.filters)
	mr.filtersMu.RUnlock()
	
	// Apply filters in priority order
	for _, filter := range filters {
		if !filter.Filter(msg) {
			mr.logger.Debug("Message filtered",
				zap.String("message_id", msg.ID),
				zap.String("filter_id", filter.ID()),
			)
			return false
		}
	}
	
	return true
}

// removeFilterByID removes filter by ID (must be called with filtersMu held)
func (mr *messageRouter) removeFilterByID(filterID string) bool {
	for i, filter := range mr.filters {
		if filter.ID() == filterID {
			mr.filters = append(mr.filters[:i], mr.filters[i+1:]...)
			return true
		}
	}
	return false
}

// sortFilters sorts filters by priority (highest first)
func (mr *messageRouter) sortFilters() {
	// Simple bubble sort by priority
	n := len(mr.filters)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if mr.filters[j].Priority() < mr.filters[j+1].Priority() {
				mr.filters[j], mr.filters[j+1] = mr.filters[j+1], mr.filters[j]
			}
		}
	}
}

// cacheCleaner periodically cleans up expired cache entries
func (mr *messageRouter) cacheCleaner() {
	ticker := time.NewTicker(mr.cacheTTL / 4) // Clean up every quarter of TTL
	defer ticker.Stop()
	
	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.cleanupExpiredEntries()
		}
	}
}

// cleanupExpiredEntries removes expired entries from cache
func (mr *messageRouter) cleanupExpiredEntries() {
	mr.cacheMu.Lock()
	defer mr.cacheMu.Unlock()
	
	now := time.Now()
	expiredCount := 0
	
	for id, entry := range mr.messageCache {
		if now.Sub(entry.timestamp) > mr.cacheTTL {
			delete(mr.messageCache, id)
			expiredCount++
		}
	}
	
	if expiredCount > 0 {
		mr.logger.Debug("Cleaned up expired cache entries",
			zap.Int("expired_count", expiredCount),
			zap.Int("remaining_count", len(mr.messageCache)),
		)
	}
}

// cleanupOldEntries removes oldest entries from cache (must be called with cacheMu held)
func (mr *messageRouter) cleanupOldEntries(removeCount int) {
	if removeCount <= 0 {
		return
	}
	
	// Find oldest entries
	type cacheItem struct {
		id        string
		timestamp time.Time
	}
	
	items := make([]cacheItem, 0, len(mr.messageCache))
	for id, entry := range mr.messageCache {
		items = append(items, cacheItem{id: id, timestamp: entry.timestamp})
	}
	
	// Sort by timestamp (oldest first)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].timestamp.After(items[j].timestamp) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	
	// Remove oldest entries
	removed := 0
	for i := 0; i < len(items) && removed < removeCount; i++ {
		delete(mr.messageCache, items[i].id)
		removed++
	}
	
	mr.logger.Debug("Cleaned up old cache entries",
		zap.Int("removed_count", removed),
		zap.Int("remaining_count", len(mr.messageCache)),
	)
}

// MessageFilter implementation
func (mf *messageFilter) ID() string {
	return mf.id
}

func (mf *messageFilter) Priority() int {
	return mf.priority
}

func (mf *messageFilter) Filter(msg *TransportMessage) bool {
	if mf.filterFn == nil {
		return true
	}
	return mf.filterFn(msg)
}

// NewMessageFilter creates a new message filter
func NewMessageFilter(id string, priority int, filterFn func(*TransportMessage) bool) MessageFilter {
	return &messageFilter{
		id:       id,
		priority: priority,
		filterFn: filterFn,
	}
}

// Pre-defined filters

// NewSizeFilter creates filter that blocks messages larger than maxSize
func NewSizeFilter(maxSize int) MessageFilter {
	return NewMessageFilter(
		"size_filter",
		50,
		func(msg *TransportMessage) bool {
			return GetMessageSize(msg) <= maxSize
		},
	)
}

// NewTTLFilter creates filter that blocks expired messages
func NewTTLFilter() MessageFilter {
	return NewMessageFilter(
		"ttl_filter",
		100,
		func(msg *TransportMessage) bool {
			return !IsMessageExpired(msg)
		},
	)
}

// NewSenderFilter creates filter that blocks messages from specific senders
func NewSenderFilter(blockedSenders []string) MessageFilter {
	blockedSet := make(map[string]bool)
	for _, sender := range blockedSenders {
		blockedSet[sender] = true
	}
	
	return NewMessageFilter(
		"sender_filter",
		75,
		func(msg *TransportMessage) bool {
			return !blockedSet[msg.Sender]
		},
	)
}