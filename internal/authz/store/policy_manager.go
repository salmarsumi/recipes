package store

import (
	"context"

	"github.com/salmarsumi/recipes/internal/authz"
)

// PolicyManager defines the operations needed to manage the policy store.
type PolicyManager[TGroupId any, TPermissionId any, TUserId any] interface {
	UpdateGroupPermissions(ctx context.Context, groupId TGroupId, permissions []TPermissionId) error
	UpdateGroupUsers(ctx context.Context, groupId TGroupId, users []TUserId) error
	UpdateUserGroups(ctx context.Context, userId TUserId, groups []TGroupId) error
	CreateGroup(ctx context.Context, groupId TGroupId, groupName string) error
	DeleteGroup(ctx context.Context, groupId TGroupId) error
	ChangeGroupName(ctx context.Context, groupId TGroupId, newGroupName string) error
	DeleteUser(ctx context.Context, userId TUserId) error
	ReadPolicy(ctx context.Context) (*authz.Policy, error)
}
