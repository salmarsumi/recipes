package store

type ErrorCode int

const (
	DefaultError ErrorCode = iota
	Concurrency
	GroupNotFound
	DuplicateGroupName
	NoUserRecordsDeleted
)

type ErrordDescription string

const (
	defaultErrorDescription         = "An unknown failure has occurred"
	concurrencyDescription          = "The operation failed due to a concurrency issue"
	groupNotFoundDescription        = "The group was not found"
	duplicateGroupNameDescription   = "The group name already exists"
	noUserRecordsDeletedDescription = "No user records were deleted"
)

// PolicyError represents an error that occurred during the policy store operations.
type PolicyStoreError struct {
	Code        ErrorCode
	Description ErrordDescription
}

// Error returns the description of the PolicyStoreError.
// It implements the error interface.
func (e *PolicyStoreError) Error() string {
	return string(e.Description)
}

func defaultError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        DefaultError,
		Description: defaultErrorDescription,
	}
}

func concurrencyError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        Concurrency,
		Description: concurrencyDescription,
	}
}

func groupNotFoundError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        GroupNotFound,
		Description: groupNotFoundDescription,
	}
}

func duplicateGroupNameError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        DuplicateGroupName,
		Description: duplicateGroupNameDescription,
	}
}

func noUserRecordsDeletedError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        NoUserRecordsDeleted,
		Description: noUserRecordsDeletedDescription,
	}
}
