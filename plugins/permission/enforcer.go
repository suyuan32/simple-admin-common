// Copyright 2026 The Ryan SU Authors (https://github.com/suyuan32). All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package permission provides on-demand API permission checks for Simple
// Admin. It stores policies in sys_permissions instead of loading them into
// process memory.
package permission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/suyuan32/simple-admin-common/config"
)

const (
	tableName        = "sys_permissions"
	legacyTableName  = "casbin_rules"
	legacyPolicyType = "p"
)

var keyMatch2Parameter = regexp.MustCompile(`:[^/]+`)

// Policy is the API permission row stored in sys_permissions. Its field mapping
// deliberately preserves the legacy layout: Subject -> v0, Object -> v1,
// Action -> v2 and Domain -> v3. v4 and v5 remain reserved for future use.
type Policy struct {
	Subject string
	Object  string
	Action  string
	Domain  string
}

// Enforcer evaluates Simple Admin's API permission policy directly from the
// database. It does not keep all policies in memory and does not subscribe to
// policy reload notifications.
//
// The supported matcher is the project's current matcher:
// r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act.
type Enforcer struct {
	db      *sql.DB
	dialect string
}

// New creates an on-demand permission enforcer from an existing database pool.
// The caller owns db and is responsible for closing it. Call InitDatabase from
// the service's database initialization flow before serving protected routes.
func New(db *sql.DB, dialect string) (*Enforcer, error) {
	if db == nil {
		return nil, errors.New("permission database cannot be nil")
	}

	return &Enforcer{
		db:      db,
		dialect: strings.ToLower(dialect),
	}, nil
}

// NewWithDatabaseConf opens a dedicated database pool for permission checks.
func NewWithDatabaseConf(conf config.DatabaseConf) (*Enforcer, error) {
	db, err := conf.NewDB()
	if err != nil {
		return nil, err
	}

	enforcer, err := New(db, conf.Type)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return enforcer, nil
}

// MustNewWithDatabaseConf is NewWithDatabaseConf that panics on configuration
// or database errors, matching the existing service-context construction style.
func MustNewWithDatabaseConf(conf config.DatabaseConf) *Enforcer {
	enforcer, err := NewWithDatabaseConf(conf)
	if err != nil {
		panic(err)
	}

	return enforcer
}

// Close closes the database pool used by this enforcer. Do not call Close when
// the pool passed to New is shared by another component.
func (e *Enforcer) Close() error {
	return e.db.Close()
}

// InitDatabase creates and indexes the permission table. It is idempotent and
// deliberately separate from New so permission storage is maintained by the
// application's database initialization flow rather than by an ORM adapter.
func (e *Enforcer) InitDatabase(ctx context.Context) error {
	return e.ensureSchema(ctx)
}

// MigrateFromCasbinRules creates sys_permissions and copies legacy p policies
// from casbin_rules only when sys_permissions did not exist before the call.
// An existing table is always authoritative, including an intentionally empty
// one, so later initialization cannot restore revoked permissions.
//
// The legacy mapping is v0..v5 -> v0..v5. Casbin g rules and all other ptype
// values are ignored because Simple Admin's API authorization only uses p. An
// optional timeout of zero disables the migration deadline. A failed first
// migration removes the newly created target table so it can be retried after
// legacy data is corrected.
func (e *Enforcer) MigrateFromCasbinRules(ctx context.Context, timeouts ...time.Duration) (migrated int64, err error) {
	migrationCtx, cancel, err := withMigrationTimeout(ctx, timeouts)
	if err != nil {
		return 0, err
	}
	defer cancel()

	targetExists, err := e.tableExists(migrationCtx, tableName)
	if err != nil {
		return 0, err
	}
	if targetExists {
		return 0, nil
	}

	if err := e.InitDatabase(migrationCtx); err != nil {
		return 0, err
	}
	createdTarget := true
	defer func() {
		if !createdTarget || err == nil {
			return
		}
		if dropErr := e.dropTable(context.Background(), tableName); dropErr != nil {
			err = fmt.Errorf("%w; failed to remove incomplete %s table: %v", err, tableName, dropErr)
		}
	}()

	legacyExists, err := e.tableExists(migrationCtx, legacyTableName)
	if err != nil || !legacyExists {
		return 0, err
	}

	tx, err := e.db.BeginTx(migrationCtx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	countQuery := fmt.Sprintf(
		"SELECT COUNT(1) FROM %s WHERE ptype = %s",
		legacyTableName, e.placeholder(1),
	)
	var sourceCount int64
	if err = tx.QueryRowContext(migrationCtx, countQuery, legacyPolicyType).Scan(&sourceCount); err != nil {
		return 0, err
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (v0, v1, v2, v3, v4, v5) "+
			"SELECT v0, v1, v2, v3, v4, v5 FROM %s WHERE ptype = %s",
		tableName, legacyTableName, e.placeholder(1),
	)
	if _, err = tx.ExecContext(migrationCtx, insertQuery, legacyPolicyType); err != nil {
		return 0, err
	}

	countQuery = fmt.Sprintf("SELECT COUNT(1) FROM %s", tableName)
	var targetCount int64
	if err = tx.QueryRowContext(migrationCtx, countQuery).Scan(&targetCount); err != nil {
		return 0, err
	}
	if targetCount != sourceCount {
		return 0, fmt.Errorf("permission migration row count mismatch: casbin_rules=%d, sys_permissions=%d", sourceCount, targetCount)
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return targetCount, nil
}

func withMigrationTimeout(ctx context.Context, timeouts []time.Duration) (context.Context, context.CancelFunc, error) {
	if len(timeouts) > 1 {
		return nil, nil, errors.New("permission migration accepts at most one timeout")
	}
	if len(timeouts) == 0 || timeouts[0] == 0 {
		return ctx, func() {}, nil
	}
	if timeouts[0] < 0 {
		return nil, nil, errors.New("permission migration timeout cannot be negative")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeouts[0])
	return timeoutCtx, cancel, nil
}

// Check returns whether any subject has permission to access object with action.
// Static policies use the unique index (v0, v1, v2, ...).
// Dynamic keyMatch2 policies are only queried when the exact lookup misses.
// The optional domain is stored in v3. Omitting it preserves non-tenant
// behavior, while tenant middleware must pass its trusted tenant ID.
func (e *Enforcer) Check(ctx context.Context, subjects []string, object, action string, domains ...string) (bool, error) {
	domain, err := resolveDomain(domains)
	if err != nil {
		return false, err
	}

	subjects = uniqueNonEmpty(subjects)
	if len(subjects) == 0 {
		return false, nil
	}

	matched, err := e.hasExactPolicy(ctx, subjects, object, action, domain)
	if err != nil || matched {
		return matched, err
	}

	return e.hasMatchedPatternPolicy(ctx, subjects, object, action, domain)
}

// BatchEnforce keeps the familiar Casbin method shape for gradual migrations.
// It accepts [subject, object, action] and [subject, object, action, domain]
// requests and evaluates them on demand.
// New middleware should prefer Check so multiple roles are queried together.
func (e *Enforcer) BatchEnforce(requests [][]any) ([]bool, error) {
	results := make([]bool, len(requests))
	for i, request := range requests {
		if len(request) != 3 && len(request) != 4 {
			return nil, fmt.Errorf("permission request %d must contain subject, object, action and optional domain", i)
		}

		subject, ok := request[0].(string)
		if !ok {
			return nil, fmt.Errorf("permission request %d subject must be a string", i)
		}
		object, ok := request[1].(string)
		if !ok {
			return nil, fmt.Errorf("permission request %d object must be a string", i)
		}
		action, ok := request[2].(string)
		if !ok {
			return nil, fmt.Errorf("permission request %d action must be a string", i)
		}

		var domain string
		if len(request) == 4 {
			domain, ok = request[3].(string)
			if !ok {
				return nil, fmt.Errorf("permission request %d domain must be a string", i)
			}
		}

		allowed, err := e.Check(context.Background(), []string{subject}, object, action, domain)
		if err != nil {
			return nil, err
		}
		results[i] = allowed
	}

	return results, nil
}

// ListPolicies returns a role's API policies without loading the full model.
// The optional domain scopes the query to v3; its default is the non-tenant
// empty domain.
func (e *Enforcer) ListPolicies(ctx context.Context, subject string, domains ...string) ([]Policy, error) {
	domain, err := resolveDomain(domains)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT v0, v1, v2, v3 FROM %s WHERE v0 = %s AND v3 = %s ORDER BY id",
		tableName, e.placeholder(1), e.placeholder(2),
	)
	rows, err := e.db.QueryContext(ctx, query, subject, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]Policy, 0)
	for rows.Next() {
		var policy Policy
		if err = rows.Scan(&policy.Subject, &policy.Object, &policy.Action, &policy.Domain); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, rows.Err()
}

// ReplacePolicies atomically replaces one subject's API policies in one domain.
// Omitting the optional domain preserves the non-tenant v3 = ” behavior.
func (e *Enforcer) ReplacePolicies(ctx context.Context, subject string, policies []Policy, domains ...string) error {
	if subject == "" {
		return errors.New("permission subject cannot be empty")
	}
	domain, err := resolveDomain(domains)
	if err != nil {
		return err
	}

	policies = normalizePolicies(subject, domain, policies)
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	deleteQuery := fmt.Sprintf(
		"DELETE FROM %s WHERE v0 = %s AND v3 = %s",
		tableName, e.placeholder(1), e.placeholder(2),
	)
	if _, err = tx.ExecContext(ctx, deleteQuery, subject, domain); err != nil {
		return err
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (v0, v1, v2, v3, v4, v5) VALUES (%s, %s, %s, %s, %s, %s)",
		tableName,
		e.placeholder(1), e.placeholder(2), e.placeholder(3), e.placeholder(4),
		e.placeholder(5), e.placeholder(6),
	)
	for _, policy := range policies {
		if _, err = tx.ExecContext(ctx, insertQuery,
			subject, policy.Object, policy.Action, policy.Domain, "", ""); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemovePolicies removes one subject's API policies in one domain. It returns
// whether rows were removed.
func (e *Enforcer) RemovePolicies(ctx context.Context, subject string, domains ...string) (bool, error) {
	domain, err := resolveDomain(domains)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE v0 = %s AND v3 = %s",
		tableName, e.placeholder(1), e.placeholder(2),
	)
	result, err := e.db.ExecContext(ctx, query, subject, domain)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// RenameSubject keeps policies consistent after a role code changes. The
// optional domain prevents identical role codes in other tenants being renamed.
func (e *Enforcer) RenameSubject(ctx context.Context, oldSubject, newSubject string, domains ...string) error {
	if oldSubject == "" || newSubject == "" {
		return errors.New("permission subject cannot be empty")
	}
	if oldSubject == newSubject {
		return nil
	}
	domain, err := resolveDomain(domains)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"UPDATE %s SET v0 = %s WHERE v0 = %s AND v3 = %s",
		tableName, e.placeholder(1), e.placeholder(2), e.placeholder(3),
	)
	_, err = e.db.ExecContext(ctx, query, newSubject, oldSubject, domain)
	return err
}

// AddPolicies adds only policies that do not already exist. Each policy must
// include Subject; Domain is optional for the non-tenant model.
func (e *Enforcer) AddPolicies(ctx context.Context, policies []Policy) (bool, error) {
	policies = normalizePolicyRows(policies)
	if len(policies) == 0 {
		return false, nil
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (v0, v1, v2, v3, v4, v5) "+
			"SELECT %s, %s, %s, %s, %s, %s WHERE NOT EXISTS "+
			"(SELECT 1 FROM %s WHERE v0 = %s AND v1 = %s AND v2 = %s AND v3 = %s AND v4 = %s AND v5 = %s)",
		tableName,
		e.placeholder(1), e.placeholder(2), e.placeholder(3), e.placeholder(4), e.placeholder(5), e.placeholder(6),
		tableName,
		e.placeholder(7), e.placeholder(8), e.placeholder(9), e.placeholder(10), e.placeholder(11), e.placeholder(12),
	)

	added := false
	for _, policy := range policies {
		result, err := tx.ExecContext(ctx, insertQuery,
			policy.Subject, policy.Object, policy.Action, policy.Domain, "", "",
			policy.Subject, policy.Object, policy.Action, policy.Domain, "", "")
		if err != nil {
			return false, err
		}
		if affected, err := result.RowsAffected(); err == nil && affected > 0 {
			added = true
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return added, nil
}

// RemoveDomainPolicies removes all API permissions that belong to one domain.
// It is used when a tenant is disabled, reauthorized, or deleted.
func (e *Enforcer) RemoveDomainPolicies(ctx context.Context, domain string) (bool, error) {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE v3 = %s",
		tableName, e.placeholder(1),
	)
	result, err := e.db.ExecContext(ctx, query, domain)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RemovePoliciesByResources removes permissions for the supplied API resources
// in only the specified domains, without enumerating all policies in memory.
func (e *Enforcer) RemovePoliciesByResources(ctx context.Context, domains []string, resources []Policy) (bool, error) {
	domains = uniqueNonEmpty(domains)
	resources = normalizeResources(resources)
	if len(domains) == 0 || len(resources) == 0 {
		return false, nil
	}

	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE v3 IN (%s) AND v1 = %s AND v2 = %s",
		tableName, e.placeholders(1, len(domains)),
		e.placeholder(len(domains)+1), e.placeholder(len(domains)+2),
	)
	removed := false
	for _, resource := range resources {
		args := make([]any, 0, len(domains)+2)
		args = append(args, stringsToAny(domains)...)
		args = append(args, resource.Object, resource.Action)
		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return false, err
		}
		if affected, err := result.RowsAffected(); err == nil && affected > 0 {
			removed = true
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return removed, nil
}

func (e *Enforcer) hasExactPolicy(ctx context.Context, subjects []string, object, action, domain string) (bool, error) {
	args := make([]any, 0, len(subjects)+3)
	args = append(args, stringsToAny(subjects)...)
	args = append(args, object, action, domain)

	query := fmt.Sprintf(
		"SELECT 1 FROM %s WHERE v0 IN (%s) AND v1 = %s AND v2 = %s AND v3 = %s LIMIT 1",
		tableName, e.placeholders(1, len(subjects)),
		e.placeholder(len(subjects)+1), e.placeholder(len(subjects)+2), e.placeholder(len(subjects)+3),
	)
	var one int
	err := e.db.QueryRowContext(ctx, query, args...).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (e *Enforcer) hasMatchedPatternPolicy(ctx context.Context, subjects []string, object, action, domain string) (bool, error) {
	args := make([]any, 0, len(subjects)+4)
	args = append(args, stringsToAny(subjects)...)
	args = append(args, action, domain, "%*%", "%:%")

	query := fmt.Sprintf(
		"SELECT v1 FROM %s WHERE v0 IN (%s) AND v2 = %s AND v3 = %s AND (v1 LIKE %s OR v1 LIKE %s)",
		tableName, e.placeholders(1, len(subjects)),
		e.placeholder(len(subjects)+1), e.placeholder(len(subjects)+2), e.placeholder(len(subjects)+3), e.placeholder(len(subjects)+4),
	)
	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var pattern string
		if err = rows.Scan(&pattern); err != nil {
			return false, err
		}
		if keyMatch2(object, pattern) {
			return true, nil
		}
	}

	return false, rows.Err()
}

func (e *Enforcer) placeholder(index int) string {
	if e.dialect == "postgres" {
		return fmt.Sprintf("$%d", index)
	}

	return "?"
}

func (e *Enforcer) placeholders(start, count int) string {
	values := make([]string, 0, count)
	for i := 0; i < count; i++ {
		values = append(values, e.placeholder(start+i))
	}

	return strings.Join(values, ", ")
}

func (e *Enforcer) ensureSchema(ctx context.Context) error {
	var createTable string
	switch e.dialect {
	case "mysql":
		createTable = `CREATE TABLE IF NOT EXISTS sys_permissions (
			id BIGINT NOT NULL AUTO_INCREMENT,
			v0 VARCHAR(24) NOT NULL DEFAULT '',
			v1 VARCHAR(512) NOT NULL DEFAULT '',
			v2 VARCHAR(24) NOT NULL DEFAULT '',
			v3 VARCHAR(32) NOT NULL DEFAULT '',
			v4 VARCHAR(12) NOT NULL DEFAULT '',
			v5 VARCHAR(12) NOT NULL DEFAULT '',
			PRIMARY KEY (id),
			UNIQUE KEY sys_permissions_v0_v1_v2_v3_v4_v5 (v0, v1, v2, v3, v4, v5),
			KEY sys_permissions_v3_v0_v2 (v3, v0, v2),
			KEY sys_permissions_v3_v1_v2 (v3, v1, v2)
		)`
	case "postgres":
		createTable = `CREATE TABLE IF NOT EXISTS sys_permissions (
			id BIGSERIAL PRIMARY KEY,
			v0 VARCHAR(24) NOT NULL DEFAULT '',
			v1 VARCHAR(512) NOT NULL DEFAULT '',
			v2 VARCHAR(24) NOT NULL DEFAULT '',
			v3 VARCHAR(32) NOT NULL DEFAULT '',
			v4 VARCHAR(12) NOT NULL DEFAULT '',
			v5 VARCHAR(12) NOT NULL DEFAULT '',
			CONSTRAINT sys_permissions_v0_v1_v2_v3_v4_v5 UNIQUE (v0, v1, v2, v3, v4, v5)
		)`
	case "sqlite", "sqlite3":
		createTable = `CREATE TABLE IF NOT EXISTS sys_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			v0 TEXT NOT NULL DEFAULT '' CHECK (length(v0) <= 24),
			v1 TEXT NOT NULL DEFAULT '' CHECK (length(v1) <= 512),
			v2 TEXT NOT NULL DEFAULT '' CHECK (length(v2) <= 24),
			v3 TEXT NOT NULL DEFAULT '' CHECK (length(v3) <= 32),
			v4 TEXT NOT NULL DEFAULT '' CHECK (length(v4) <= 12),
			v5 TEXT NOT NULL DEFAULT '' CHECK (length(v5) <= 12),
			CONSTRAINT sys_permissions_v0_v1_v2_v3_v4_v5 UNIQUE (v0, v1, v2, v3, v4, v5)
		)`
	default:
		return fmt.Errorf("unsupported permission database type %q", e.dialect)
	}

	if _, err := e.db.ExecContext(ctx, createTable); err != nil {
		return err
	}

	if err := e.ensureLookupIndex(ctx, "sys_permissions_v3_v0_v2", "v3, v0, v2"); err != nil {
		return err
	}
	return e.ensureLookupIndex(ctx, "sys_permissions_v3_v1_v2", "v3, v1, v2")
}

func (e *Enforcer) ensureLookupIndex(ctx context.Context, indexName, columns string) error {
	var query string
	switch e.dialect {
	case "mysql":
		var exists int
		query = "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?"
		if err := e.db.QueryRowContext(ctx, query, tableName, indexName).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return nil
		}
		query = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, columns)
	case "postgres":
		query = fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, columns)
	case "sqlite", "sqlite3":
		query = fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, columns)
	default:
		return fmt.Errorf("unsupported permission database type %q", e.dialect)
	}

	_, err := e.db.ExecContext(ctx, query)
	return err
}

func (e *Enforcer) dropTable(ctx context.Context, name string) error {
	_, err := e.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", name))
	return err
}

func (e *Enforcer) tableExists(ctx context.Context, name string) (bool, error) {
	var query string
	switch e.dialect {
	case "mysql":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	case "postgres":
		query = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1"
	case "sqlite", "sqlite3":
		query = "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?"
	default:
		return false, fmt.Errorf("unsupported permission database type %q", e.dialect)
	}

	var count int
	if err := e.db.QueryRowContext(ctx, query, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func normalizePolicies(subject, domain string, policies []Policy) []Policy {
	seen := make(map[string]struct{}, len(policies))
	result := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		if policy.Object == "" || policy.Action == "" {
			continue
		}
		policy.Subject = subject
		policy.Domain = domain
		key := policy.Object + "\x00" + policy.Action + "\x00" + policy.Domain
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, policy)
	}

	return result
}

func normalizePolicyRows(policies []Policy) []Policy {
	seen := make(map[string]struct{}, len(policies))
	result := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		if policy.Subject == "" || policy.Object == "" || policy.Action == "" {
			continue
		}
		key := policy.Subject + "\x00" + policy.Object + "\x00" + policy.Action + "\x00" + policy.Domain
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, policy)
	}
	return result
}

func normalizeResources(resources []Policy) []Policy {
	seen := make(map[string]struct{}, len(resources))
	result := make([]Policy, 0, len(resources))
	for _, resource := range resources {
		if resource.Object == "" || resource.Action == "" {
			continue
		}
		key := resource.Object + "\x00" + resource.Action
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, Policy{Object: resource.Object, Action: resource.Action})
	}
	return result
}

func resolveDomain(domains []string) (string, error) {
	if len(domains) == 0 {
		return "", nil
	}
	if len(domains) != 1 {
		return "", errors.New("permission accepts at most one domain")
	}
	return domains[0], nil
}

func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for i := range values {
		result[i] = values[i]
	}

	return result
}

func keyMatch2(object, pattern string) bool {
	pattern = strings.ReplaceAll(pattern, "/*", "/.*")
	pattern = keyMatch2Parameter.ReplaceAllString(pattern, "[^/]+")
	matcher, err := regexp.Compile("^" + pattern + "$")
	return err == nil && matcher.MatchString(object)
}
