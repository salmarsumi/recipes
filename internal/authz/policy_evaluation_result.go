package authz

// Represents the result of a policy evaluation for a specific user.
type PolicyEvaluationResult struct {

	// The groups that the user is a member of.
	Groups []string

	// The permissions that the user has.
	Permissions []string
}

// Creates a new instance of PolicyEvaluationResult.
func NewPolicyEvaluationResult(groups []string, permissions []string) *PolicyEvaluationResult {
	return &PolicyEvaluationResult{Groups: groups, Permissions: permissions}
}
