package authz

import (
	"errors"
	"slices"
)

// Represents a single users group in the system with all the users
// that are members of that specific group.
// Given a user the group instance can evaluate whether this user
// is a member of the specified group.
type Group struct {
	Name  string
	Users []string
}

// NewGroup creates a new Group with the specified name and list of users.
// Parameters:
//   - name: The name of the group.
//   - users: A slice of strings representing the users in the group.
//
// Returns:
//
//	A pointer to the newly created Group.
func NewGroup(name string, users []string) *Group {
	return &Group{Name: name, Users: users}
}

// Evaluate checks if a given user is part of the group.
// It returns true if the user is found in the group's user list, otherwise false.
// If the provided user string is empty, it returns an error indicating that the group name is empty.
//
// Parameters:
//
//	user - the username to be checked within the group.
//
// Returns:
//
//	bool - true if the user is in the group, false otherwise.
//	error - an error if the user string is empty.
func (group *Group) Evaluate(user string) (bool, error) {
	if user == "" {
		return false, errors.New("user is empty")
	}

	return slices.Contains(group.Users, user), nil
}
