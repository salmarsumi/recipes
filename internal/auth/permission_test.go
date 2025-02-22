package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPermission calls auth.NewPermission with a name and groups, checking for a valid returned permission.
func TestNewPermission(t *testing.T) {
	name := "name"
	groups := []string{"group 1", "group 2"}

	permission := NewPermission(name, groups)

	assert.NotNil(t, permission)
	assert.Equal(t, name, permission.Name)
	assert.Equal(t, groups, permission.Groups)
}

// TestEvaluate_Error_NilGroups calls permission.Evaluate with nil groups, checking for an error.
func TestEvaluate_Error_NilGroups(t *testing.T) {
	name := "name"
	groups := []string{"group 1", "group 2"}

	permission := NewPermission(name, groups)

	isGranted, err := permission.Evaluate(nil)
	assert.Error(t, err)
	assert.False(t, isGranted)
}

// TestEvaluate_False_EmptyGroups calls permission.Evaluate with an empty groups slice, checking for a false result.
func TestEvaluate_False_EmptyGroups(t *testing.T) {
	name := "name"
	groups := []string{"group 1", "group 2"}

	permission := NewPermission(name, groups)

	isGranted, err := permission.Evaluate([]string{})
	assert.NoError(t, err)
	assert.False(t, isGranted)
}

// TestEvaluate_True_GroupsGranted calls permission.Evaluate with groups that have been granted the permission, checking for a true result.
func TestEvaluate_True_GroupsGranted(t *testing.T) {
	name := "name"
	groups := []string{"group 1", "group 2"}

	permission := NewPermission(name, groups)

	isGranted, err := permission.Evaluate([]string{"group 1"})
	assert.NoError(t, err)
	assert.True(t, isGranted)
}

// TestEvaluate_False_GroupsNotGranted calls permission.Evaluate with groups that have not been granted the permission, checking for a false result.
func TestEvaluate_False_GroupsNotGranted(t *testing.T) {
	name := "name"
	groups := []string{"group 1", "group 2"}

	permission := NewPermission(name, groups)

	isGranted, err := permission.Evaluate([]string{"group 3"})
	assert.NoError(t, err)
	assert.False(t, isGranted)
}
