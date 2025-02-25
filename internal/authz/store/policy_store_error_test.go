package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyStoreError_Error(t *testing.T) {
	tests := []struct {
		name                string
		err                 *PolicyStoreError
		expectedMsg         string
		expectedDescription ErrordDescription
		expectedCode        ErrorCode
	}{
		{
			name:                "DefaultError",
			err:                 defaultError(),
			expectedMsg:         string(defaultErrorDescription),
			expectedDescription: defaultErrorDescription,
			expectedCode:        DefaultError,
		},
		{
			name:                "ConcurrencyError",
			err:                 concurrencyError(),
			expectedMsg:         string(concurrencyDescription),
			expectedDescription: concurrencyDescription,
			expectedCode:        Concurrency,
		},
		{
			name:                "GroupNotFoundError",
			err:                 groupNotFoundError(),
			expectedMsg:         string(groupNotFoundDescription),
			expectedDescription: groupNotFoundDescription,
			expectedCode:        GroupNotFound,
		},
		{
			name:                "DuplicateGroupNameError",
			err:                 duplicateGroupNameError(),
			expectedMsg:         string(duplicateGroupNameDescription),
			expectedDescription: duplicateGroupNameDescription,
			expectedCode:        DuplicateGroupName,
		},
		{
			name:                "NoUserRecordsDeletedError",
			err:                 noUserRecordsDeletedError(),
			expectedMsg:         string(noUserRecordsDeletedDescription),
			expectedDescription: noUserRecordsDeletedDescription,
			expectedCode:        NoUserRecordsDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.err.Code)
			assert.Equal(t, tt.expectedDescription, tt.err.Description)
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
		})
	}
}
