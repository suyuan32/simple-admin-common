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

// Package permission provides on-demand API permission checks for the Simple
// Admin Casbin data model. It intentionally queries casbin_rules instead of
// loading all policies into process memory.
package permission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/suyuan32/simple-admin-common/config"
)

const (
	policyType = "p"
	tableName  = "casbin_rules"
)

var keyMatch2Parameter = regexp.MustCompile(`:[^/]+`)

// Policy is the API permission row stored in casbin_rules.
// It maps to Casbin's p = sub, obj, act definition.
type Policy struct {
	Subject string
	Object  string
	Action  string
}

// Enforcer evaluates Simple Admin's p = sub, obj, act policy directly from
// the database. Unlike casbin.Enforcer, it does not keep all policies in
// memory and it does not subscribe to policy reload notifications.
//
// The supported matcher is the project's current matcher:
// r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act.
type Enforcer struct {
	db      *sql.DB
	dialect string
}

// New creates an on-demand permission enforcer from an existing database pool.
// The caller owns db and is responsible for closing it.
func New(db *sql.DB, dialect string) (*Enforcer, error) {
	if db == nil {
		return nil, errors.New("permission database cannot be nil")
	}

	enforcer := &Enforcer{
		db:      db,
		dialect: strings.ToLower(dialect),
	}
	if err := enforcer.ensureSchema(context.Background()); err != nil {
		return nil, err
	}

	return enforcer, nil
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

// Check returns whether any subject has permission to access object with action.
// Static policies use the existing unique index (ptype, v0, v1, v2, ...).
// Dynamic keyMatch2 policies are only queried when the exact lookup misses.
func (e *Enforcer) Check(ctx context.Context, subjects []string, object, action string) (bool, error) {
	subjects = uniqueNonEmpty(subjects)
	if len(subjects) == 0 {
		return false, nil
	}

	matched, err := e.hasExactPolicy(ctx, subjects, object, action)
	if err != nil || matched {
		return matched, err
	}

	return e.hasMatchedPatternPolicy(ctx, subjects, object, action)
}

// BatchEnforce keeps the familiar Casbin method shape for gradual migrations.
// It accepts [subject, object, action] requests and evaluates them on demand.
// New middleware should prefer Check so multiple roles are queried together.
func (e *Enforcer) BatchEnforce(requests [][]any) ([]bool, error) {
	results := make([]bool, len(requests))
	for i, request := range requests {
		if len(request) != 3 {
			return nil, fmt.Errorf("permission request %d must contain subject, object and action", i)
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

		allowed, err := e.Check(context.Background(), []string{subject}, object, action)
		if err != nil {
			return nil, err
		}
		results[i] = allowed
	}

	return results, nil
}

// ListPolicies returns a role's API policies without loading the full model.
func (e *Enforcer) ListPolicies(ctx context.Context, subject string) ([]Policy, error) {
	query := fmt.Sprintf(
		"SELECT v0, v1, v2 FROM %s WHERE ptype = %s AND v0 = %s ORDER BY id",
		tableName, e.placeholder(1), e.placeholder(2),
	)
	rows, err := e.db.QueryContext(ctx, query, policyType, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]Policy, 0)
	for rows.Next() {
		var policy Policy
		if err = rows.Scan(&policy.Subject, &policy.Object, &policy.Action); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, rows.Err()
}

// ReplacePolicies atomically replaces one subject's API policies.
func (e *Enforcer) ReplacePolicies(ctx context.Context, subject string, policies []Policy) error {
	if subject == "" {
		return errors.New("permission subject cannot be empty")
	}

	policies = normalizePolicies(subject, policies)
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	deleteQuery := fmt.Sprintf(
		"DELETE FROM %s WHERE ptype = %s AND v0 = %s",
		tableName, e.placeholder(1), e.placeholder(2),
	)
	if _, err = tx.ExecContext(ctx, deleteQuery, policyType, subject); err != nil {
		return err
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (ptype, v0, v1, v2, v3, v4, v5) VALUES (%s, %s, %s, %s, %s, %s, %s)",
		tableName,
		e.placeholder(1), e.placeholder(2), e.placeholder(3), e.placeholder(4),
		e.placeholder(5), e.placeholder(6), e.placeholder(7),
	)
	for _, policy := range policies {
		if _, err = tx.ExecContext(ctx, insertQuery,
			policyType, subject, policy.Object, policy.Action, "", "", ""); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemovePolicies removes one subject's API policies. It returns whether rows
// were removed.
func (e *Enforcer) RemovePolicies(ctx context.Context, subject string) (bool, error) {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE ptype = %s AND v0 = %s",
		tableName, e.placeholder(1), e.placeholder(2),
	)
	result, err := e.db.ExecContext(ctx, query, policyType, subject)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// RenameSubject keeps policies consistent after a role code changes.
func (e *Enforcer) RenameSubject(ctx context.Context, oldSubject, newSubject string) error {
	if oldSubject == "" || newSubject == "" {
		return errors.New("permission subject cannot be empty")
	}
	if oldSubject == newSubject {
		return nil
	}

	query := fmt.Sprintf(
		"UPDATE %s SET v0 = %s WHERE ptype = %s AND v0 = %s",
		tableName, e.placeholder(1), e.placeholder(2), e.placeholder(3),
	)
	_, err := e.db.ExecContext(ctx, query, newSubject, policyType, oldSubject)
	return err
}

func (e *Enforcer) hasExactPolicy(ctx context.Context, subjects []string, object, action string) (bool, error) {
	args := make([]any, 0, len(subjects)+3)
	args = append(args, policyType)
	args = append(args, stringsToAny(subjects)...)
	args = append(args, object, action)

	query := fmt.Sprintf(
		"SELECT 1 FROM %s WHERE ptype = %s AND v0 IN (%s) AND v1 = %s AND v2 = %s LIMIT 1",
		tableName, e.placeholder(1), e.placeholders(2, len(subjects)),
		e.placeholder(len(subjects)+2), e.placeholder(len(subjects)+3),
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

func (e *Enforcer) hasMatchedPatternPolicy(ctx context.Context, subjects []string, object, action string) (bool, error) {
	args := make([]any, 0, len(subjects)+4)
	args = append(args, policyType)
	args = append(args, stringsToAny(subjects)...)
	args = append(args, action, "%*%", "%:%")

	query := fmt.Sprintf(
		"SELECT v1 FROM %s WHERE ptype = %s AND v0 IN (%s) AND v2 = %s AND (v1 LIKE %s OR v1 LIKE %s)",
		tableName, e.placeholder(1), e.placeholders(2, len(subjects)),
		e.placeholder(len(subjects)+2), e.placeholder(len(subjects)+3), e.placeholder(len(subjects)+4),
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
		createTable = `CREATE TABLE IF NOT EXISTS casbin_rules (
			id BIGINT NOT NULL AUTO_INCREMENT,
			ptype VARCHAR(24) NOT NULL DEFAULT '',
			v0 VARCHAR(24) NOT NULL DEFAULT '',
			v1 VARCHAR(255) NOT NULL DEFAULT '',
			v2 VARCHAR(24) NOT NULL DEFAULT '',
			v3 VARCHAR(32) NOT NULL DEFAULT '',
			v4 VARCHAR(12) NOT NULL DEFAULT '',
			v5 VARCHAR(12) NOT NULL DEFAULT '',
			PRIMARY KEY (id),
			UNIQUE KEY casbinrule_ptype_v0_v1_v2_v3_v4_v5 (ptype, v0, v1, v2, v3, v4, v5)
		)`
	case "postgres":
		createTable = `CREATE TABLE IF NOT EXISTS casbin_rules (
			id BIGSERIAL PRIMARY KEY,
			ptype VARCHAR(24) NOT NULL DEFAULT '',
			v0 VARCHAR(24) NOT NULL DEFAULT '',
			v1 VARCHAR(255) NOT NULL DEFAULT '',
			v2 VARCHAR(24) NOT NULL DEFAULT '',
			v3 VARCHAR(32) NOT NULL DEFAULT '',
			v4 VARCHAR(12) NOT NULL DEFAULT '',
			v5 VARCHAR(12) NOT NULL DEFAULT '',
			CONSTRAINT casbinrule_ptype_v0_v1_v2_v3_v4_v5 UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
		)`
	case "sqlite3":
		createTable = `CREATE TABLE IF NOT EXISTS casbin_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ptype TEXT NOT NULL DEFAULT '',
			v0 TEXT NOT NULL DEFAULT '',
			v1 TEXT NOT NULL DEFAULT '',
			v2 TEXT NOT NULL DEFAULT '',
			v3 TEXT NOT NULL DEFAULT '',
			v4 TEXT NOT NULL DEFAULT '',
			v5 TEXT NOT NULL DEFAULT '',
			UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
		)`
	default:
		return fmt.Errorf("unsupported permission database type %q", e.dialect)
	}

	_, err := e.db.ExecContext(ctx, createTable)
	return err
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

func normalizePolicies(subject string, policies []Policy) []Policy {
	seen := make(map[string]struct{}, len(policies))
	result := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		if policy.Object == "" || policy.Action == "" {
			continue
		}
		policy.Subject = subject
		key := policy.Object + "\x00" + policy.Action
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, policy)
	}

	return result
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
