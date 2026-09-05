package permission

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func newTestEnforcer(t *testing.T) *Enforcer {
	t.Helper()
	enforcer := newUninitializedTestEnforcer(t)
	require.NoError(t, enforcer.InitDatabase(context.Background()))
	return enforcer
}

func newUninitializedTestEnforcer(t *testing.T) *Enforcer {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	enforcer, err := New(db, "sqlite3")
	require.NoError(t, err)
	return enforcer
}

func TestEnforcerCheck(t *testing.T) {
	enforcer := newTestEnforcer(t)
	ctx := context.Background()
	require.NoError(t, enforcer.ReplacePolicies(ctx, "admin", []Policy{
		{Object: "/api/user/list", Action: "POST"},
		{Object: "/api/user/:id", Action: "GET"},
		{Object: "/api/report/*", Action: "GET"},
	}))

	allowed, err := enforcer.Check(ctx, []string{"guest", "admin", "admin"}, "/api/user/list", "POST")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = enforcer.Check(ctx, []string{"admin"}, "/api/user/42", "GET")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = enforcer.Check(ctx, []string{"admin"}, "/api/report/2026/09", "GET")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = enforcer.Check(ctx, []string{"admin"}, "/api/unknown/list", "GET")
	require.NoError(t, err)
	require.False(t, allowed)

	allowed, err = enforcer.Check(ctx, nil, "/api/user/list", "POST")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestEnforcerPolicyManagement(t *testing.T) {
	enforcer := newTestEnforcer(t)
	ctx := context.Background()

	require.NoError(t, enforcer.ReplacePolicies(ctx, "001", []Policy{
		{Object: "/first", Action: "GET"},
		{Object: "/first", Action: "GET"},
		{Object: "", Action: "POST"},
	}))
	policies, err := enforcer.ListPolicies(ctx, "001")
	require.NoError(t, err)
	require.Equal(t, []Policy{{Subject: "001", Object: "/first", Action: "GET"}}, policies)

	require.NoError(t, enforcer.RenameSubject(ctx, "001", "admin"))
	allowed, err := enforcer.Check(ctx, []string{"admin"}, "/first", "GET")
	require.NoError(t, err)
	require.True(t, allowed)

	removed, err := enforcer.RemovePolicies(ctx, "admin")
	require.NoError(t, err)
	require.True(t, removed)
	allowed, err = enforcer.Check(ctx, []string{"admin"}, "/first", "GET")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestEnforcerBatchEnforceCompatibility(t *testing.T) {
	enforcer := newTestEnforcer(t)
	require.NoError(t, enforcer.ReplacePolicies(context.Background(), "001", []Policy{
		{Object: "/health", Action: "GET"},
	}))

	results, err := enforcer.BatchEnforce([][]any{
		{"001", "/health", "GET"},
		{"001", "/health", "POST"},
	})
	require.NoError(t, err)
	require.Equal(t, []bool{true, false}, results)

	_, err = enforcer.BatchEnforce([][]any{{"001", "/health"}})
	require.Error(t, err)
}

func TestEnforcerTenantDomainIsolation(t *testing.T) {
	enforcer := newTestEnforcer(t)
	ctx := context.Background()
	require.NoError(t, enforcer.ReplacePolicies(ctx, "001", []Policy{
		{Object: "/api/order", Action: "GET"},
	}, "100"))
	require.NoError(t, enforcer.ReplacePolicies(ctx, "001", []Policy{
		{Object: "/api/user", Action: "GET"},
	}, "200"))

	allowed, err := enforcer.Check(ctx, []string{"001"}, "/api/order", "GET", "100")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = enforcer.Check(ctx, []string{"001"}, "/api/order", "GET", "200")
	require.NoError(t, err)
	require.False(t, allowed)

	policies, err := enforcer.ListPolicies(ctx, "001", "100")
	require.NoError(t, err)
	require.Equal(t, []Policy{{Subject: "001", Object: "/api/order", Action: "GET", Domain: "100"}}, policies)

	added, err := enforcer.AddPolicies(ctx, []Policy{
		{Subject: "002", Object: "/api/report", Action: "GET", Domain: "100"},
	})
	require.NoError(t, err)
	require.True(t, added)

	removed, err := enforcer.RemovePoliciesByResources(ctx, []string{"100"}, []Policy{
		{Object: "/api/order", Action: "GET"},
	})
	require.NoError(t, err)
	require.True(t, removed)

	allowed, err = enforcer.Check(ctx, []string{"001"}, "/api/order", "GET", "100")
	require.NoError(t, err)
	require.False(t, allowed)

	removed, err = enforcer.RemoveDomainPolicies(ctx, "200")
	require.NoError(t, err)
	require.True(t, removed)
	allowed, err = enforcer.Check(ctx, []string{"001"}, "/api/user", "GET", "200")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestMigrateFromCasbinRules(t *testing.T) {
	enforcer := newUninitializedTestEnforcer(t)
	ctx := context.Background()

	_, err := enforcer.db.ExecContext(ctx, `CREATE TABLE casbin_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ptype TEXT NOT NULL DEFAULT '',
		v0 TEXT NOT NULL DEFAULT '',
		v1 TEXT NOT NULL DEFAULT '',
		v2 TEXT NOT NULL DEFAULT '',
		v3 TEXT NOT NULL DEFAULT '',
		v4 TEXT NOT NULL DEFAULT '',
		v5 TEXT NOT NULL DEFAULT ''
	)`)
	require.NoError(t, err)

	_, err = enforcer.db.ExecContext(ctx, `INSERT INTO casbin_rules (ptype, v0, v1, v2, v3, v4, v5) VALUES
		('p', '001', '/api/user', 'GET', '100', '', ''),
		('p', '002', '/api/order/:id', 'GET', '200', '', ''),
		('g', 'user-1', '001', '', '', '', '')`)
	require.NoError(t, err)

	migrated, err := enforcer.MigrateFromCasbinRules(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 2, migrated)

	allowed, err := enforcer.Check(ctx, []string{"001"}, "/api/user", "GET", "100")
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = enforcer.Check(ctx, []string{"002"}, "/api/order/42", "GET", "200")
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, err = enforcer.Check(ctx, []string{"user-1"}, "001", "", "")
	require.NoError(t, err)
	require.False(t, allowed)

	migrated, err = enforcer.MigrateFromCasbinRules(ctx)
	require.NoError(t, err)
	require.Zero(t, migrated)

	rows, err := enforcer.db.QueryContext(ctx, "PRAGMA table_info(sys_permissions)")
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultV   any
			primary    int
		)
		require.NoError(t, rows.Scan(&cid, &name, &columnType, &notNull, &defaultV, &primary))
		require.NotEqual(t, "ptype", name)
	}
	require.NoError(t, rows.Err())
}

func TestMigrateFromCasbinRulesInitializesEmptyTargetWithoutLegacy(t *testing.T) {
	enforcer := newUninitializedTestEnforcer(t)
	ctx := context.Background()

	migrated, err := enforcer.MigrateFromCasbinRules(ctx)
	require.NoError(t, err)
	require.Zero(t, migrated)

	targetExists, err := enforcer.tableExists(ctx, tableName)
	require.NoError(t, err)
	require.True(t, targetExists)
	legacyExists, err := enforcer.tableExists(ctx, legacyTableName)
	require.NoError(t, err)
	require.False(t, legacyExists)

	var count int
	require.NoError(t, enforcer.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM sys_permissions").Scan(&count))
	require.Zero(t, count)
}

func TestMigrateFromCasbinRulesSkipsExistingTarget(t *testing.T) {
	enforcer := newTestEnforcer(t)
	ctx := context.Background()

	_, err := enforcer.db.ExecContext(ctx, `CREATE TABLE casbin_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ptype TEXT NOT NULL DEFAULT '',
		v0 TEXT NOT NULL DEFAULT '',
		v1 TEXT NOT NULL DEFAULT '',
		v2 TEXT NOT NULL DEFAULT '',
		v3 TEXT NOT NULL DEFAULT '',
		v4 TEXT NOT NULL DEFAULT '',
		v5 TEXT NOT NULL DEFAULT ''
	)`)
	require.NoError(t, err)
	_, err = enforcer.db.ExecContext(ctx, "INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES ('p', '001', '/api/user', 'GET')")
	require.NoError(t, err)

	migrated, err := enforcer.MigrateFromCasbinRules(ctx)
	require.NoError(t, err)
	require.Zero(t, migrated)

	allowed, err := enforcer.Check(ctx, []string{"001"}, "/api/user", "GET")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestMigrateFromCasbinRulesRemovesIncompleteTarget(t *testing.T) {
	enforcer := newUninitializedTestEnforcer(t)
	ctx := context.Background()

	_, err := enforcer.db.ExecContext(ctx, `CREATE TABLE casbin_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ptype TEXT NOT NULL DEFAULT '',
		v0 TEXT NOT NULL DEFAULT '',
		v1 TEXT NOT NULL DEFAULT '',
		v2 TEXT NOT NULL DEFAULT '',
		v3 TEXT NOT NULL DEFAULT '',
		v4 TEXT NOT NULL DEFAULT '',
		v5 TEXT NOT NULL DEFAULT ''
	)`)
	require.NoError(t, err)
	_, err = enforcer.db.ExecContext(ctx,
		"INSERT INTO casbin_rules (ptype, v0, v1, v2) VALUES (?, ?, ?, ?)",
		"p", "001", strings.Repeat("/", 513), "GET",
	)
	require.NoError(t, err)

	_, err = enforcer.MigrateFromCasbinRules(ctx)
	require.Error(t, err)
	targetExists, err := enforcer.tableExists(ctx, tableName)
	require.NoError(t, err)
	require.False(t, targetExists)

	_, err = enforcer.db.ExecContext(ctx, "UPDATE casbin_rules SET v1 = '/api/user' WHERE v0 = '001'")
	require.NoError(t, err)
	migrated, err := enforcer.MigrateFromCasbinRules(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, migrated)
}

func TestMigrationTimeout(t *testing.T) {
	ctx, cancel, err := withMigrationTimeout(context.Background(), []time.Duration{0})
	require.NoError(t, err)
	defer cancel()
	_, hasDeadline := ctx.Deadline()
	require.False(t, hasDeadline)

	ctx, cancel, err = withMigrationTimeout(context.Background(), []time.Duration{time.Second})
	require.NoError(t, err)
	defer cancel()
	_, hasDeadline = ctx.Deadline()
	require.True(t, hasDeadline)

	_, _, err = withMigrationTimeout(context.Background(), []time.Duration{-time.Second})
	require.Error(t, err)
	_, _, err = withMigrationTimeout(context.Background(), []time.Duration{time.Second, time.Second})
	require.Error(t, err)
}

func TestPermissionSchemaLengthsAndGlobalUniqueIndex(t *testing.T) {
	enforcer := newTestEnforcer(t)
	ctx := context.Background()
	insert := "INSERT INTO sys_permissions (v0, v1, v2, v3, v4, v5) VALUES (?, ?, ?, ?, ?, ?)"
	valid := []any{"001", "/api/user", "GET", "100", "", ""}

	_, err := enforcer.db.ExecContext(ctx, insert, valid...)
	require.NoError(t, err)
	_, err = enforcer.db.ExecContext(ctx, insert, valid...)
	require.Error(t, err, "v0 through v5 must be globally unique as one policy")
	_, err = enforcer.db.ExecContext(ctx, insert, "002", strings.Repeat("/", 512), "GET", "100", "", "")
	require.NoError(t, err, "a 512-character path must be accepted")

	testCases := []struct {
		name  string
		value []any
	}{
		{name: "v0", value: []any{strings.Repeat("r", 25), "/api/user/0", "GET", "100", "", ""}},
		{name: "v1", value: []any{"002", strings.Repeat("/", 513), "GET", "100", "", ""}},
		{name: "v2", value: []any{"003", "/api/user/2", strings.Repeat("M", 25), "100", "", ""}},
		{name: "v3", value: []any{"004", "/api/user/3", "GET", strings.Repeat("1", 33), "", ""}},
		{name: "v4", value: []any{"005", "/api/user/4", "GET", "100", strings.Repeat("x", 13), ""}},
		{name: "v5", value: []any{"006", "/api/user/5", "GET", "100", "", strings.Repeat("x", 13)}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := enforcer.db.ExecContext(ctx, insert, testCase.value...)
			require.Error(t, err)
		})
	}
}
