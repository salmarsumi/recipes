package auth

import (
	"errors"

	"github.com/salmarsumi/recipes/internal/shared"
)

// Represents the entire policy configuration with all permissions and user groups defined in the system.
// This class will be the single source of truth regarding which user can have what permission.Given a user
// the policy instance can evaluate and return what permissions and membership the user has.
type Policy struct {
	Permissions []Permission
	Groups      []Group
}

// NewPolicy creates a new Policy instance with the specified permissions and groups.
func NewPolicy(permissions []Permission, groups []Group) *Policy {
	return &Policy{Permissions: permissions, Groups: groups}
}

// Evaluate assesses the given user's permissions based on the policy.
// It returns a PolicyEvaluationResult which indicates whether the user
// meets the policy requirements, and an error if the evaluation fails.
//
// Parameters:
//
//	user - the username to be evaluated against the policy.
//
// Returns:
//
//	*PolicyEvaluationResult - the result of the policy evaluation.
//	error - an error if the evaluation process encounters an issue.
func (policy *Policy) Evaluate(user string) (*PolicyEvaluationResult, error) {
	if user == "" {
		return nil, errors.New("user is empty")
	}

	// get the user groups
	groups := shared.Filter(policy.Groups, func(group Group) bool {
		result, error := group.Evaluate(user)
		if error != nil {
			return result
		}
		return false
	}, func(group Group) string {
		return group.Name
	})

	permissions := shared.Filter(policy.Permissions, func(permission Permission) bool {
		result, error := permission.Evaluate(groups)
		if error != nil {
			return result
		}
		return false
	}, func(permission Permission) string {
		return permission.Name
	})

	return &PolicyEvaluationResult{Groups: groups, Permissions: permissions}, nil
}
