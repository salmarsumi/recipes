package store

type ErrorCode int

const (
	DefaultError ErrorCode = iota
	Concurrency
	GroupNotFound
	DuplicateGroupName
	NoUserRecordsDeleted
	DatabaseError
)

type ErrordDescription string

const (
	defaultErrorDescription         = "An unknown failure has occurred"
	concurrencyDescription          = "The operation failed due to a concurrency issue"
	groupNotFoundDescription        = "The group was not found"
	duplicateGroupNameDescription   = "The group name already exists"
	noUserRecordsDeletedDescription = "No user records were deleted"
	databaseErrorDescription        = "An error occurred while interacting with the database"
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

func NewDefaultError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        DefaultError,
		Description: defaultErrorDescription,
	}
}

func NewConcurrencyError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        Concurrency,
		Description: concurrencyDescription,
	}
}

func NewGroupNotFoundError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        GroupNotFound,
		Description: groupNotFoundDescription,
	}
}

func NewDuplicateGroupNameError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        DuplicateGroupName,
		Description: duplicateGroupNameDescription,
	}
}

func NewNoUserRecordsDeletedError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        NoUserRecordsDeleted,
		Description: noUserRecordsDeletedDescription,
	}
}

func NewDataBaseError() *PolicyStoreError {
	return &PolicyStoreError{
		Code:        DatabaseError,
		Description: databaseErrorDescription,
	}
}
