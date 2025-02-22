package auth

import (
	"errors"
	"slices"
)

// Represents a single users group in the system with all the users
// that are members of that specific group.
// Given a user the group instance can evaluate whether this user
// is a member of the specified group.
type Group struct {

	// The Name of the group.
	Name string

	// The collection of Users that are members of the group.
	Users []string
}

// Creates a new instance of Group.
func NewGroup(name string, users []string) *Group {
	return &Group{Name: name, Users: users}
}

// Evaluate whether a user is a member of the current group.
func (group *Group) Evaluate(user string) (bool, error) {
	if user == "" {
		return false, errors.New("group name is empty")
	}

	return slices.Contains(group.Users, user), nil
}
