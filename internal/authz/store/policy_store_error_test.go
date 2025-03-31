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
			err:                 NewDefaultError(),
			expectedMsg:         string(defaultErrorDescription),
			expectedDescription: defaultErrorDescription,
			expectedCode:        DefaultError,
		},
		{
			name:                "ConcurrencyError",
			err:                 NewConcurrencyError(),
			expectedMsg:         string(concurrencyDescription),
			expectedDescription: concurrencyDescription,
			expectedCode:        Concurrency,
		},
		{
			name:                "GroupNotFoundError",
			err:                 NewGroupNotFoundError(),
			expectedMsg:         string(groupNotFoundDescription),
			expectedDescription: groupNotFoundDescription,
			expectedCode:        GroupNotFound,
		},
		{
			name:                "DuplicateGroupNameError",
			err:                 NewNameExistsError(),
			expectedMsg:         string(nameAlreadyExistsDescription),
			expectedDescription: nameAlreadyExistsDescription,
			expectedCode:        NameAlreadyExist,
		},
		{
			name:                "NoUserRecordsDeletedError",
			err:                 NewNoUserRecordsDeletedError(),
			expectedMsg:         string(noUserRecordsDeletedDescription),
			expectedDescription: noUserRecordsDeletedDescription,
			expectedCode:        NoUserRecordsDeleted,
		},
		{
			name:                "DataBaseError",
			err:                 NewDataBaseError(),
			expectedMsg:         string(databaseErrorDescription),
			expectedDescription: databaseErrorDescription,
			expectedCode:        DatabaseError,
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
