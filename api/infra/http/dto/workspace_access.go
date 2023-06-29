package dto

type WorkspaceAccess struct {
	WorkspaceName string `json:"workspaceName"`
	Owner         bool   `json:"owner"`
	User          bool   `json:"user"`
}
