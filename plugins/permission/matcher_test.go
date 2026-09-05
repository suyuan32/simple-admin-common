package permission

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyMatch2Compatibility(t *testing.T) {
	testCases := []struct {
		object  string
		pattern string
		matched bool
	}{
		{object: "/resource1", pattern: "/:resource", matched: true},
		{object: "/", pattern: "/:resource", matched: false},
		{object: "/foo/bar", pattern: "/foo/*", matched: true},
		{object: "/foo", pattern: "/foo/*", matched: false},
		{object: "/foo", pattern: "/foo*", matched: true},
		{object: "/foo/bar", pattern: "/foo/:id", matched: true},
		{object: "/foo/bar/baz", pattern: "/foo/:id", matched: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.pattern, func(t *testing.T) {
			require.Equal(t, testCase.matched, keyMatch2(testCase.object, testCase.pattern))
		})
	}
}

func TestPolicyNormalization(t *testing.T) {
	policies := normalizePolicies("001", []Policy{
		{Object: "/user/list", Action: "POST"},
		{Object: "/user/list", Action: "POST"},
		{Object: "", Action: "POST"},
		{Object: "/user/get", Action: ""},
	})
	require.Equal(t, []Policy{{Subject: "001", Object: "/user/list", Action: "POST"}}, policies)
	require.Equal(t, []string{"001", "002"}, uniqueNonEmpty([]string{"001", "", "001", "002"}))
}

func TestPostgresPlaceholders(t *testing.T) {
	enforcer := &Enforcer{dialect: "postgres"}
	require.Equal(t, "$3", enforcer.placeholder(3))
	require.Equal(t, "$2, $3, $4", enforcer.placeholders(2, 3))
}
