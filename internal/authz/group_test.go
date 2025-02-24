package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewGroup calls auth.NewGroup with a name and users, checking for a valid returned group.
func TestNewGroup(t *testing.T) {
	name := "name"
	users := []string{"user 1", "user 2"}

	group := NewGroup(name, users)

	assert.NotNil(t, group)
	assert.Equal(t, name, group.Name)
	assert.Equal(t, users, group.Users)
}

// TestEvaluate_Error_EmptyUser calls group.Evaluate with an empty user, checking for an error.
func TestEvaluate_Error_EmptyUser(t *testing.T) {
	name := "name"
	users := []string{"user 1", "user 2"}

	group := NewGroup(name, users)

	isMember, err := group.Evaluate("")
	assert.Error(t, err)
	assert.False(t, isMember)
}

// TestEvaluate_True_UserInGroup calls group.Evaluate with a user that is in the group, checking for a true result.
func TestEvaluate_True_UserInGroup(t *testing.T) {
	name := "name"
	users := []string{"user 1", "user 2"}

	group := NewGroup(name, users)

	isMember, err := group.Evaluate("user 1")
	assert.NoError(t, err)
	assert.True(t, isMember)
}

// TestEvaluate_False_UserNotInGroup calls group.Evaluate with a user that is not in the group, checking for a false result.
func TestEvaluate_False_UserNotInGroup(t *testing.T) {
	name := "name"
	users := []string{"user 1", "user 2"}

	group := NewGroup(name, users)

	isMember, err := group.Evaluate("user 3")
	assert.NoError(t, err)
	assert.False(t, isMember)
}
