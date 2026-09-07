package permission

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultPermissionCacheTTL       = 24 * time.Hour
	defaultPermissionLocalCacheTTL  = 2 * time.Second
	defaultPermissionLocalCacheSize = 4096
	defaultPermissionCachePrefix    = "SIMPLE:PERMISSION:"
	permissionCacheVersionFormat    = "scoped-v2"
	permissionPatternCacheTTL       = 30 * time.Second
	permissionPatternCacheSize      = 1024
	permissionPatternCacheMaxRules  = 4096
)

// PermissionCacheOptions controls the optional Redis/L1 permission decision
// cache. RedisTTL controls retention of decision entries; policy mutations
// invalidate affected Redis decisions through global, tenant, or role
// version keys.
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

func (e *Enforcer) invalidateCache(scopes ...permissionCacheScope) error {
	var err error
	if e.cache != nil {
		err = e.cache.invalidate(scopes...)
	}
	if e.patternCache != nil {
		e.patternCache.clear(scopes...)
	}
	return err
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
	subjects  []string
	domain    string
	expiresAt time.Time
}

type permissionCacheScope struct {
	tenantID   string
	roleCode   string
	tenantWide bool
}

func permissionRoleCacheScope(tenantID, roleCode string) permissionCacheScope {
	return permissionCacheScope{tenantID: tenantID, roleCode: roleCode}
}

func permissionTenantCacheScope(tenantID string) permissionCacheScope {
	return permissionCacheScope{tenantID: tenantID, tenantWide: true}
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
	parts := make([]string, 0, len(subjects)+3)
	parts = append(parts, domain, action, object)
	parts = append(parts, permissionCanonicalSubjects(subjects)...)
	return c.keyPrefix + "DECISION:" + permissionCacheHash(parts...)
}

// get returns a decision and the Redis version used to validate a miss. A
// local hit intentionally avoids Redis for low-latency hot routes; its short
// TTL bounds cross-instance staleness.
func (c *permissionDecisionCache) get(ctx context.Context, key string, subjects []string, domain string) (allowed, hit bool, version string, versionKnown bool) {
	if allowed, ok := c.getLocal(key); ok {
		c.localHits.Add(1)
		return allowed, true, "", false
	}

	versionKeys := c.versionKeys(subjects, domain)
	keys := make([]string, 0, len(versionKeys)+1)
	keys = append(keys, versionKeys...)
	keys = append(keys, key)
	values, err := c.rds.MGet(ctx, keys...).Result()
	if err != nil {
		c.redisErrors.Add(1)
		return false, false, "", false
	}

	versions := make([]string, len(versionKeys))
	for index := range versionKeys {
		versions[index] = "0"
		if index < len(values) {
			if value, ok := redisValueString(values[index]); ok && value != "" {
				versions[index] = value
			}
		}
	}
	version = permissionCacheVersion(versions...)
	decisionIndex := len(versionKeys)
	if len(values) <= decisionIndex {
		c.redisMisses.Add(1)
		return false, false, version, true
	}

	value, ok := redisValueString(values[decisionIndex])
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
	c.setLocal(key, cachedAllowed, subjects, domain)
	return cachedAllowed, true, version, true
}

func (c *permissionDecisionCache) set(ctx context.Context, key string, subjects []string, domain string, allowed bool, version string) {
	c.setLocal(key, allowed, subjects, domain)
	if version == "" {
		return
	}
	if err := c.rds.Set(ctx, key, encodePermissionDecision(version, allowed), c.redisTTL).Err(); err == nil {
		c.redisSets.Add(1)
	} else {
		c.redisErrors.Add(1)
	}
}

func (c *permissionDecisionCache) invalidate(scopes ...permissionCacheScope) error {
	c.clearLocal(scopes...)
	if c.rds == nil {
		return nil
	}

	keys := c.invalidationVersionKeys(scopes)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		_, lastErr = c.rds.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for _, key := range keys {
				pipe.Incr(ctx, key)
			}
			return nil
		})
		if lastErr == nil {
			return nil
		}
		if attempt == 2 {
			break
		}

		timer := time.NewTimer(time.Duration(25*(1<<attempt)) * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			attempt = 2
		case <-timer.C:
		}
	}

	c.redisErrors.Add(1)
	return lastErr
}

func (c *permissionDecisionCache) versionKeys(subjects []string, domain string) []string {
	canonicalSubjects := permissionCanonicalSubjects(subjects)
	keys := make([]string, 0, len(canonicalSubjects)+2)
	keys = append(keys, c.versionKey, c.tenantVersionKey(domain))
	for _, subject := range canonicalSubjects {
		keys = append(keys, c.roleVersionKey(domain, subject))
	}
	return keys
}

func (c *permissionDecisionCache) invalidationVersionKeys(scopes []permissionCacheScope) []string {
	if len(scopes) == 0 {
		return []string{c.versionKey}
	}

	unique := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		key := c.roleVersionKey(scope.tenantID, scope.roleCode)
		if scope.tenantWide {
			key = c.tenantVersionKey(scope.tenantID)
		}
		unique[key] = struct{}{}
	}
	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *permissionDecisionCache) tenantVersionKey(tenantID string) string {
	return c.keyPrefix + "VERSION:TENANT:" + permissionCacheKeyParts(tenantID)
}

func (c *permissionDecisionCache) roleVersionKey(tenantID, roleCode string) string {
	return c.keyPrefix + "VERSION:ROLE:" + permissionCacheKeyParts(tenantID, roleCode)
}

func permissionCanonicalSubjects(subjects []string) []string {
	canonicalSubjects := append([]string(nil), subjects...)
	sort.Strings(canonicalSubjects)
	return canonicalSubjects
}

func permissionCacheHash(values ...string) string {
	hash := sha256.New()
	for _, value := range values {
		_, _ = fmt.Fprintf(hash, "%d:", len(value))
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func permissionCacheVersion(versions ...string) string {
	parts := make([]string, 0, len(versions)+1)
	parts = append(parts, permissionCacheVersionFormat)
	parts = append(parts, versions...)
	return permissionCacheHash(parts...)
}

func permissionCacheKeyParts(values ...string) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = strconv.Itoa(len(value)) + ":" + value
	}
	return strings.Join(parts, ":")
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

func (c *permissionDecisionCache) setLocal(key string, allowed bool, subjects []string, domain string) {
	if c.localTTL <= 0 || c.maxLocalEntries <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &permissionLocalCacheEntry{
		key:       key,
		allowed:   allowed,
		subjects:  permissionCanonicalSubjects(subjects),
		domain:    domain,
		expiresAt: time.Now().Add(c.localTTL),
	}
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

func (c *permissionDecisionCache) clearLocal(scopes ...permissionCacheScope) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(scopes) > 0 {
		for key, element := range c.local {
			entry := element.Value.(*permissionLocalCacheEntry)
			if !permissionCacheEntryMatches(entry.domain, entry.subjects, scopes) {
				continue
			}
			delete(c.local, key)
			c.lru.Remove(element)
		}
		return
	}

	c.local = make(map[string]*list.Element)
	c.lru.Init()
}

func permissionCacheEntryMatches(domain string, subjects []string, scopes []permissionCacheScope) bool {
	for _, scope := range scopes {
		if scope.tenantID != domain {
			continue
		}
		if scope.tenantWide {
			return true
		}
		if slices.Contains(subjects, scope.roleCode) {
			return true
		}
	}
	return false
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

type permissionPatternCache struct {
	mu      sync.Mutex
	entries map[string]*list.Element
	lru     *list.List
}

type permissionPatternCacheEntry struct {
	key       string
	patterns  []permissionPattern
	version   string
	subjects  []string
	domain    string
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

func (c *permissionPatternCache) set(key string, patterns []permissionPattern, version string, subjects []string, domain string) {
	if len(patterns) > permissionPatternCacheMaxRules {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	entry := &permissionPatternCacheEntry{
		key:       key,
		patterns:  append([]permissionPattern(nil), patterns...),
		version:   version,
		subjects:  permissionCanonicalSubjects(subjects),
		domain:    domain,
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

func (c *permissionPatternCache) clear(scopes ...permissionCacheScope) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(scopes) > 0 {
		for key, element := range c.entries {
			entry := element.Value.(*permissionPatternCacheEntry)
			if !permissionCacheEntryMatches(entry.domain, entry.subjects, scopes) {
				continue
			}
			delete(c.entries, key)
			c.lru.Remove(element)
		}
		return
	}

	c.entries = make(map[string]*list.Element, permissionPatternCacheSize)
	c.lru.Init()
}
