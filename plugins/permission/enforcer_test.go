package permission

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func newTestEnforcer(t *testing.T) *Enforcer {
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
