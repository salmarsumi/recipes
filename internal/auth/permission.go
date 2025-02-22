package auth

import (
	"errors"
)

// Represents a single system permission with all the
// groups assigned that specific permission.Given a
// collection of groups the permission instance can
// evaluate whether these groups have been granted
// the specified permission.
type Permission struct {
	Name   string
	Groups []string
}

// NewPermission creates a new Permission instance with the specified name and groups.
//
// Parameters:
//   - name: The name of the permission.
//   - groups: A slice of strings representing the groups associated with the permission.
//
// Returns:
//
//	A pointer to a Permission instance initialized with the provided name and groups.
func NewPermission(name string, groups []string) *Permission {
	return &Permission{Name: name, Groups: groups}
}

// Evaluate whether a collection of groups are assigned the current permission.
// Evaluate checks if the permission is granted based on the provided groups.
// It returns true if the permission is granted, otherwise false.
// An error is returned if the evaluation process encounters any issues.
//
// Parameters:
//
//	groups []string - A slice of group names to evaluate against the permission.
//
// Returns:
//
//	bool - True if the permission is granted, otherwise false.
//	error - An error if the evaluation process fails.
func (permission *Permission) Evaluate(groups []string) (bool, error) {
	if groups == nil {
		return false, errors.New("groups is nil")
	}

	if len(groups) == 0 {
		return false, nil
	}

	// use a map for faster lookup
	groupsMap := make(map[string]struct{}, len(permission.Groups))
	for _, group := range permission.Groups {
		groupsMap[group] = struct{}{}
	}

	// check if the groups intersect
	for _, group := range groups {
		if _, exists := groupsMap[group]; exists {
			return true, nil
		}
	}

	return false, nil
}
