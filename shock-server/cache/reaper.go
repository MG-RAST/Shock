package cache

import (
	"regexp"
	"strconv"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
)

var (
	CacheReaper  *CacheCleanup
	TTLRegex     = regexp.MustCompile(`^(\d+)(M|H|D)$`)
)

// CacheCleanup handles periodic cleanup of expired cache items
type CacheCleanup struct {
	ttl time.Duration
}

// InitCacheReaper initializes the cache reaper if TTL is configured
func InitCacheReaper() {
	if conf.PATH_CACHE == "" {
		logger.Info("(InitCacheReaper) Cache path not configured, reaper disabled")
		return
	}

	if conf.CACHE_TTL == "" {
		logger.Info("(InitCacheReaper) Cache TTL not configured, reaper disabled")
		return
	}

	ttl := ParseTTL(conf.CACHE_TTL)
	if ttl == 0 {
		logger.Errorf("(InitCacheReaper) Invalid CACHE_TTL format: %s (expected format: 24H, 7D, 30M)", conf.CACHE_TTL)
		return
	}

	CacheReaper = &CacheCleanup{
		ttl: ttl,
	}

	logger.Infof("(InitCacheReaper) CacheReaper started with TTL: %s", conf.CACHE_TTL)
}

// Handle runs the cache cleanup loop
func (cr *CacheCleanup) Handle() {
	// Run cleanup every TTL/4 interval
	waitDuration := cr.ttl / 4
	if waitDuration < time.Minute {
		waitDuration = time.Minute
	}

	for {
		time.Sleep(waitDuration)
		cr.cleanExpiredCache()
	}
}

// cleanExpiredCache removes cache items older than TTL
func (cr *CacheCleanup) cleanExpiredCache() {
	if CacheMap == nil {
		return
	}

	now := time.Now()
	expiredItems := []string{}

	// Collect expired items (read lock)
	CacheMapLock.RLock()
	for uuid, item := range CacheMap {
		// Use Access time if available, otherwise use CreatedOn
		itemTime := item.Access
		if itemTime.IsZero() {
			itemTime = item.CreatedOn
		}

		age := now.Sub(itemTime)
		if age > cr.ttl {
			expiredItems = append(expiredItems, uuid)
			logger.Debug(2, "(CacheReaper) Item %s expired (age: %s, ttl: %s)", uuid, age, cr.ttl)
		}
	}
	CacheMapLock.RUnlock()

	// Remove expired items
	for _, uuid := range expiredItems {
		err := Remove(uuid)
		if err != nil {
			logger.Errorf("(CacheReaper) Failed to remove expired item %s: %s", uuid, err.Error())
		} else {
			logger.Infof("(CacheReaper) Removed expired cache item: %s", uuid)
		}
	}

	if len(expiredItems) > 0 {
		logger.Infof("(CacheReaper) Cleaned up %d expired cache items", len(expiredItems))
	}
}

// ParseTTL parses a TTL string like "24H", "7D", "30M" into a time.Duration
func ParseTTL(ttlStr string) time.Duration {
	matches := TTLRegex.FindStringSubmatch(ttlStr)
	if matches == nil {
		return 0
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	unit := matches[2]
	switch unit {
	case "M":
		return time.Duration(value) * time.Minute
	case "H":
		return time.Duration(value) * time.Hour
	case "D":
		return time.Duration(value) * 24 * time.Hour
	default:
		return 0
	}
}
