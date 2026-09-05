package permission

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/suyuan32/simple-admin-common/config"
)

const (
	defaultPermissionCacheTTL       = 30 * time.Second
	defaultPermissionLocalCacheTTL  = 2 * time.Second
	defaultPermissionLocalCacheSize = 4096
	defaultPermissionCachePrefix    = "SIMPLE_ADMIN:PERMISSION:"
	permissionPatternCacheTTL       = 30 * time.Second
	permissionPatternCacheSize      = 1024
	permissionPatternCacheMaxRules  = 4096
)

// PermissionCacheOptions controls the optional Redis/L1 permission decision
// cache. RedisTTL is also the maximum cross-process staleness window.
type PermissionCacheOptions struct {
	RedisTTL        time.Duration
	LocalTTL        time.Duration
	MaxLocalEntries int
	KeyPrefix       string
}

// DefaultPermissionCacheOptions returns conservative cache defaults for API
// authorization. LocalTTL can be set to zero to disable the process-local L1.
func DefaultPermissionCacheOptions() PermissionCacheOptions {
	return PermissionCacheOptions{
		RedisTTL:        defaultPermissionCacheTTL,
		LocalTTL:        defaultPermissionLocalCacheTTL,
		MaxLocalEntries: defaultPermissionLocalCacheSize,
		KeyPrefix:       defaultPermissionCachePrefix,
	}
}

// WithRedisCache enables a bounded L1 cache backed by Redis. Cache failures
// are fail-open to the database: authorization does not depend on Redis
// availability.
func WithRedisCache(rds redis.UniversalClient, options ...PermissionCacheOptions) EnforcerOption {
	return func(e *Enforcer) {
		cacheOptions := DefaultPermissionCacheOptions()
		if len(options) > 0 {
			cacheOptions = options[0]
			if cacheOptions.RedisTTL <= 0 {
				cacheOptions.RedisTTL = defaultPermissionCacheTTL
			}
			if cacheOptions.MaxLocalEntries < 0 {
				cacheOptions.MaxLocalEntries = 0
			}
			if cacheOptions.KeyPrefix == "" {
				cacheOptions.KeyPrefix = defaultPermissionCachePrefix
			}
		}
		e.cache = newPermissionDecisionCache(rds, cacheOptions)
	}
}

// PermissionCacheStats exposes cache behavior for diagnostics and load tests.
type PermissionCacheStats struct {
	LocalHits   uint64
	RedisHits   uint64
	RedisMisses uint64
	RedisErrors uint64
	RedisSets   uint64
}

func (e *Enforcer) PermissionCacheStats() PermissionCacheStats {
	if e.cache == nil {
		return PermissionCacheStats{}
	}

	return e.cache.stats()
}

func (e *Enforcer) invalidateCache() {
	if e.cache != nil {
		e.cache.invalidate()
	}
	if e.patternCache != nil {
		e.patternCache.clear()
	}
}

type permissionDecisionCache struct {
	rds             redis.UniversalClient
	redisTTL        time.Duration
	localTTL        time.Duration
	maxLocalEntries int
	keyPrefix       string
	versionKey      string

	mu    sync.Mutex
	local map[string]*list.Element
	lru   *list.List

	localHits   atomic.Uint64
	redisHits   atomic.Uint64
	redisMisses atomic.Uint64
	redisErrors atomic.Uint64
	redisSets   atomic.Uint64
}

type permissionLocalCacheEntry struct {
	key       string
	allowed   bool
	expiresAt time.Time
}

func newPermissionDecisionCache(rds redis.UniversalClient, options PermissionCacheOptions) *permissionDecisionCache {
	if rds == nil {
		return nil
	}
	if options.RedisTTL <= 0 {
		options.RedisTTL = defaultPermissionCacheTTL
	}
	if options.LocalTTL < 0 {
		options.LocalTTL = 0
	}
	if options.MaxLocalEntries < 0 {
		options.MaxLocalEntries = 0
	}
	if options.KeyPrefix == "" {
		options.KeyPrefix = defaultPermissionCachePrefix
	}

	return &permissionDecisionCache{
		rds:             rds,
		redisTTL:        options.RedisTTL,
		localTTL:        options.LocalTTL,
		maxLocalEntries: options.MaxLocalEntries,
		keyPrefix:       options.KeyPrefix,
		versionKey:      options.KeyPrefix + "VERSION",
		local:           make(map[string]*list.Element),
		lru:             list.New(),
	}
}

func (c *permissionDecisionCache) key(subjects []string, object, action, domain string) string {
	canonicalSubjects := append([]string(nil), subjects...)
	sort.Strings(canonicalSubjects)

	hash := sha256.New()
	writePart := func(value string) {
		_, _ = fmt.Fprintf(hash, "%d:", len(value))
		_, _ = hash.Write([]byte(value))
	}

	writePart(domain)
	writePart(action)
	writePart(object)
	for _, subject := range canonicalSubjects {
		writePart(subject)
	}

	return c.keyPrefix + "DECISION:" + hex.EncodeToString(hash.Sum(nil))
}

// get returns a decision and the Redis version used to validate a miss. A
// local hit intentionally avoids Redis for low-latency hot routes; its short
// TTL bounds cross-instance staleness.
func (c *permissionDecisionCache) get(ctx context.Context, key string) (allowed, hit bool, version string, versionKnown bool) {
	if allowed, ok := c.getLocal(key); ok {
		c.localHits.Add(1)
		return allowed, true, "", false
	}

	values, err := c.rds.MGet(ctx, c.versionKey, key).Result()
	if err != nil {
		c.redisErrors.Add(1)
		return false, false, "", false
	}

	version = "0"
	if len(values) > 0 {
		if value, ok := redisValueString(values[0]); ok && value != "" {
			version = value
		}
	}
	if len(values) < 2 {
		c.redisMisses.Add(1)
		return false, false, version, true
	}

	value, ok := redisValueString(values[1])
	if !ok {
		c.redisMisses.Add(1)
		return false, false, version, true
	}

	cachedVersion, cachedAllowed, ok := decodePermissionDecision(value)
	if !ok || cachedVersion != version {
		c.redisMisses.Add(1)
		return false, false, version, true
	}

	c.redisHits.Add(1)
	c.setLocal(key, cachedAllowed)
	return cachedAllowed, true, version, true
}

func (c *permissionDecisionCache) set(ctx context.Context, key string, allowed bool, version string) {
	if version == "" {
		version = "0"
	}
	c.setLocal(key, allowed)
	if err := c.rds.Set(ctx, key, encodePermissionDecision(version, allowed), c.redisTTL).Err(); err == nil {
		c.redisSets.Add(1)
	}
}

func (c *permissionDecisionCache) invalidate() {
	c.clearLocal()
	if c.rds == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = c.rds.Incr(ctx, c.versionKey).Err()
}

func (c *permissionDecisionCache) getLocal(key string) (bool, bool) {
	if c.localTTL <= 0 || c.maxLocalEntries <= 0 {
		return false, false
	}

	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.local[key]
	if !ok {
		return false, false
	}
	entry := element.Value.(*permissionLocalCacheEntry)
	if !now.Before(entry.expiresAt) {
		delete(c.local, key)
		c.lru.Remove(element)
		return false, false
	}

	c.lru.MoveToFront(element)
	return entry.allowed, true
}

func (c *permissionDecisionCache) setLocal(key string, allowed bool) {
	if c.localTTL <= 0 || c.maxLocalEntries <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &permissionLocalCacheEntry{key: key, allowed: allowed, expiresAt: time.Now().Add(c.localTTL)}
	if element, ok := c.local[key]; ok {
		element.Value = entry
		c.lru.MoveToFront(element)
		return
	}

	element := c.lru.PushFront(entry)
	c.local[key] = element
	if len(c.local) <= c.maxLocalEntries {
		return
	}

	oldest := c.lru.Back()
	if oldest == nil {
		return
	}
	oldestEntry := oldest.Value.(*permissionLocalCacheEntry)
	delete(c.local, oldestEntry.key)
	c.lru.Remove(oldest)
}

func (c *permissionDecisionCache) clearLocal() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.local = make(map[string]*list.Element)
	c.lru.Init()
}

func (c *permissionDecisionCache) stats() PermissionCacheStats {
	return PermissionCacheStats{
		LocalHits:   c.localHits.Load(),
		RedisHits:   c.redisHits.Load(),
		RedisMisses: c.redisMisses.Load(),
		RedisErrors: c.redisErrors.Load(),
		RedisSets:   c.redisSets.Load(),
	}
}

func encodePermissionDecision(version string, allowed bool) string {
	return version + "\x00" + strconv.FormatBool(allowed)
}

func decodePermissionDecision(value string) (string, bool, bool) {
	parts := strings.Split(value, "\x00")
	if len(parts) != 2 || parts[0] == "" {
		return "", false, false
	}
	allowed, err := strconv.ParseBool(parts[1])
	if err != nil {
		return "", false, false
	}
	return parts[0], allowed, true
}

func redisValueString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	default:
		return "", false
	}
}

func permissionCacheNamespace(c config.DatabaseConf) string {
	identity := strings.Join([]string{
		strings.ToLower(c.Type),
		c.Host,
		strconv.Itoa(c.Port),
		c.DBName,
	}, "|")
	hash := sha256.Sum256([]byte(identity))
	return defaultPermissionCachePrefix + hex.EncodeToString(hash[:8]) + ":"
}

type permissionPatternCache struct {
	mu      sync.Mutex
	entries map[string]*list.Element
	lru     *list.List
}

type permissionPatternCacheEntry struct {
	key       string
	patterns  []permissionPattern
	version   string
	expiresAt time.Time
}

func newPermissionPatternCache() *permissionPatternCache {
	return &permissionPatternCache{
		entries: make(map[string]*list.Element, permissionPatternCacheSize),
		lru:     list.New(),
	}
}

func (c *permissionPatternCache) get(key, version string, versionKnown bool) ([]permissionPattern, bool) {
	if !versionKnown {
		return nil, false
	}

	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	entry := element.Value.(*permissionPatternCacheEntry)
	if entry.version != version || !now.Before(entry.expiresAt) {
		delete(c.entries, key)
		c.lru.Remove(element)
		return nil, false
	}
	c.lru.MoveToFront(element)
	return append([]permissionPattern(nil), entry.patterns...), true
}

func (c *permissionPatternCache) set(key string, patterns []permissionPattern, version string) {
	if len(patterns) > permissionPatternCacheMaxRules {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	entry := &permissionPatternCacheEntry{
		key:       key,
		patterns:  append([]permissionPattern(nil), patterns...),
		version:   version,
		expiresAt: time.Now().Add(permissionPatternCacheTTL),
	}
	if element, ok := c.entries[key]; ok {
		element.Value = entry
		c.lru.MoveToFront(element)
		return
	}

	element := c.lru.PushFront(entry)
	c.entries[key] = element
	if len(c.entries) <= permissionPatternCacheSize {
		return
	}
	oldest := c.lru.Back()
	if oldest == nil {
		return
	}
	oldestEntry := oldest.Value.(*permissionPatternCacheEntry)
	delete(c.entries, oldestEntry.key)
	c.lru.Remove(oldest)
}

func (c *permissionPatternCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*list.Element, permissionPatternCacheSize)
	c.lru.Init()
}
