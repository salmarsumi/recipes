package auth

// Defines the operations needed to evaluate users against a policy instance.
type PolicyOperations interface {
	Evaluate(user string) (*PolicyEvaluationResult, error)
	HasPermission(user string, permission string) (bool, error)
	IsInGroup(user string, group string) (bool, error)
}
