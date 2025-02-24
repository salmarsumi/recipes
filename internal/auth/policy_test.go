package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEvaluate_EmptyUser calls policy.Evaluate with an empty user, checking for an error.
func TestEvaluate_EmptyUser(t *testing.T) {
	policy := &Policy{}
	result, err := policy.Evaluate("")
	assert.Nil(t, result)
	assert.EqualError(t, err, "user is empty")
}

// TestEvaluate_UserWithGroupsAndPermissions calls policy.Evaluate with a user that has groups and permissions, checking for a valid result.
func TestEvaluate_UserWithGroupsAndPermissions(t *testing.T) {
	groups := []Group{
		*NewGroup("admin", []string{"adminuser"}),
		*NewGroup("reader", []string{"readeruser"}),
	}

	permissions := []Permission{
		*NewPermission("read", []string{"reader"}),
		*NewPermission("write", []string{"admin"}),
	}

	policy := NewPolicy(permissions, groups)
	readerResult, readerErr := policy.Evaluate("readeruser")
	adminResult, adminErr := policy.Evaluate("adminuser")

	assert.NoError(t, readerErr)
	assert.NotNil(t, readerResult)
	assert.Equal(t, []string{"reader"}, readerResult.Groups)
	assert.Equal(t, []string{"read"}, readerResult.Permissions)

	assert.NoError(t, adminErr)
	assert.NotNil(t, adminResult)
	assert.Equal(t, []string{"admin"}, adminResult.Groups)
	assert.Equal(t, []string{"write"}, adminResult.Permissions)
}

// TestEvaluate_UserWithoutGroups calls policy.Evaluate with a user that has no groups, checking for an empty result.
func TestEvaluate_UserWithoutGroupsAndPermissions(t *testing.T) {
	groups := []Group{
		*NewGroup("admin", []string{"adminuser"}),
	}

	permissions := []Permission{
		*NewPermission("write", []string{"admin"}),
	}

	policy := NewPolicy(permissions, groups)
	result, err := policy.Evaluate("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Groups)
	assert.Empty(t, result.Permissions)
}

// TestEvaluate_EmptyPolicy calls policy.Evaluate with an empty policy, checking for an empty result.
func TestEvaluate_EmptyPolicy(t *testing.T) {
	groups := []Group{}

	permissions := []Permission{}

	policy := NewPolicy(permissions, groups)
	result, err := policy.Evaluate("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Groups)
	assert.Empty(t, result.Permissions)
}
