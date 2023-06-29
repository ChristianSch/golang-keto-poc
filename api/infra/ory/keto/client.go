package keto

import (
	"context"
	"io"

	"github.com/ChristianSch/keto-go/api/domain"
	ory "github.com/ory/client-go"
	"go.uber.org/zap"
)

type KetoClient struct {
	readClient  *ory.APIClient
	writeClient *ory.APIClient
	logger      *zap.Logger
}

// NewKetoClient creates a new keto client
// FIXME: this should be configurable
func NewKetoClient(logger *zap.Logger) *KetoClient {
	writeConfiguration := ory.NewConfiguration()
	writeConfiguration.Servers = []ory.ServerConfiguration{
		{
			URL: "http://localhost:4467", // Write API
		},
	}

	readConfiguration := ory.NewConfiguration()
	readConfiguration.Servers = []ory.ServerConfiguration{
		{
			URL: "http://localhost:4466", // Read API
		},
	}

	return &KetoClient{
		writeClient: ory.NewAPIClient(writeConfiguration),
		readClient:  ory.NewAPIClient(readConfiguration),
		logger:      logger,
	}
}

// CheckUserAccessForWorkspaceName checks "workspace:<w>#view@user:<u>" relation
func (k *KetoClient) checkWorkspaceAccessForRelation(w string, u domain.User, relation string) (bool, error) {
	k.logger.Debug("check", zap.String("relation", relation), zap.String("workspace", w), zap.String("user", u.Name))
	check, r, err := k.readClient.PermissionApi.CheckPermission(context.Background()).
		Namespace("Workspace").
		Object(w).
		Relation(relation).
		SubjectSetObject(u.Name).
		SubjectSetRelation("").
		SubjectSetNamespace("User").Execute()

	if r.StatusCode != 200 {
		resBody, _ := io.ReadAll(r.Body)
		k.logger.Debug("check res", zap.Int("status", r.StatusCode), zap.String("body", string(resBody)))
	}

	if err != nil {
		return false, err
	}

	return check.Allowed, nil
}

// CheckUserAccessForWorkspaceName checks if a user has access to a workspace with the view permit.
// That means either owner or user.
func (k *KetoClient) CheckUserAccessForWorkspaceName(w string, u domain.User) (bool, error) {
	return k.checkWorkspaceAccessForRelation(w, u, "view")
}

// CheckUserAccessAsWorkspaceOwner checks if a user has access to a workspace as owner
func (k *KetoClient) CheckUserAccessAsWorkspaceOwner(w string, u domain.User) (bool, error) {
	return k.checkWorkspaceAccessForRelation(w, u, "owners")
}

// CheckUserAccessAsWorkspaceOwner checks if a user has access to a workspace as a user
func (k *KetoClient) CheckUserAccessAsWorkspaceUser(w string, u domain.User) (bool, error) {
	return k.checkWorkspaceAccessForRelation(w, u, "users")
}

type WorkspaceAccess struct {
	WorkspaceName string
	Owner         bool
	User          bool
}

// listWorkspaceRelationsPage lists all relations for a user for the given page token.
// If page is nil, the first page is returned.
// If the page token is empty in the response, the last page is returned.
func (k *KetoClient) listWorkspaceRelationsPage(u domain.User, page *string) (*ory.Relationships, error) {
	res, r, err := k.readClient.RelationshipApi.GetRelationships(context.Background()).
		Namespace("Workspace").
		SubjectSetNamespace("User").
		SubjectSetObject(u.Name).
		SubjectSetRelation("").
		Execute()

	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		resBody, _ := io.ReadAll(r.Body)
		k.logger.Debug("expand res", zap.Int("status", r.StatusCode), zap.String("body", string(resBody)))
	}

	return res, nil
}

// ListAccessibleWorkspaces lists all workspaces a user has access to and if they are owner or user
func (k *KetoClient) ListAccessibleWorkspaces(u domain.User) ([]WorkspaceAccess, error) {
	var out []WorkspaceAccess
	var nextPageToken *string

	// Iterate over pages. If the page token is empty, the last page is reached.
	for {
		rels, err := k.listWorkspaceRelationsPage(u, nil)
		nextPageToken = rels.NextPageToken

		for _, w := range rels.RelationTuples {
			out = append(out, WorkspaceAccess{
				WorkspaceName: w.Object,
				Owner:         w.Relation == "owners",
				User:          w.Relation == "users",
			})
		}

		if err != nil {
			return nil, err
		}

		if nextPageToken == nil || *nextPageToken == "" {
			break
		}
	}

	return out, nil
}
