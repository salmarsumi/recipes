package auth

// Represents the result of a policy evaluation for a specific user.
type PolicyEvaluationResult struct {

	// The groups that the user is a member of.
	Groups []string

	// The permissions that the user has.
	Permissions []string
}
