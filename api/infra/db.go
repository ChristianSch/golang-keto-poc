// simple in-memory database
package infra

import (
	"github.com/ChristianSch/keto-go/api/domain"
)

var workspaces = []domain.Workspace{
	{Name: "1",
		Units: []domain.Unit{
			{Name: "a"},
			{Name: "b"},
		},
	},
	{Name: "2",
		Units: []domain.Unit{
			{Name: "a"},
		},
	},
}

var users = []domain.User{
	{Name: "alice"},
	{Name: "bob"},
	{Name: "eve"},
}

// GetWorkspaces returns all workspaces
func GetWorkspaces() []domain.Workspace {
	return workspaces
}

// GetWorkspaceByName returns a workspace by name
func GetWorkspaceByName(name string) *domain.Workspace {
	for _, workspace := range workspaces {
		if workspace.Name == name {
			return &workspace
		}
	}
	return nil
}

// AddWorkspace adds a workspace
func AddWorkspace(workspace domain.Workspace) {
	workspaces = append(workspaces, workspace)
}

// GetUsers returns all users
func GetUsers() []domain.User {
	return users
}

// GetUserByName returns a user by name
func GetUserByName(name string) *domain.User {
	for _, user := range users {
		if user.Name == name {
			return &user
		}
	}

	return nil
}

// AddUser adds a user
func AddUser(user domain.User) {
	users = append(users, user)
}

// GetUnits returns all units of a workspace
func GetUnits(workspaceName string) []domain.Unit {
	workspace := GetWorkspaceByName(workspaceName)
	if workspace == nil {
		return nil
	}
	return workspace.Units
}
