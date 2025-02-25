package store

import "github.com/salmarsumi/recipes/internal/authz"

// PolicyManager defines the operations needed to manage the policy store.
type PolicyManager[TGroupId any, TPermissionId any, TUserId any] interface {
	UpdateGroupPermissions(groupId TGroupId, permissions []TPermissionId) error
	UpdateGroupUsers(groupId TGroupId, users []TUserId) error
	UpdateUserGroups(userId TUserId, groups []TGroupId) error
	CreateGroup(groupId TGroupId, groupName string) error
	DeleteGroup(groupId TGroupId) error
	ChangeGroupName(groupId TGroupId, newGroupName string) error
	DeleteUser(userId TUserId) error
	ReadPolicy() (*authz.Policy, error)
}
