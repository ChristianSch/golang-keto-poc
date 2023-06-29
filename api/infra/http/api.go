package http

import (
	"github.com/ChristianSch/keto-go/api/domain"
	"github.com/ChristianSch/keto-go/api/infra"
	"github.com/ChristianSch/keto-go/api/infra/http/dto"
	"github.com/ChristianSch/keto-go/api/infra/ory/keto"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func ApiFactory(api *fiber.App, logger *zap.Logger, ketoClient *keto.KetoClient) {
	// middleware that transforms the X-User header to a c.Local("user") value
	// FIXME: this obviously should be replaced with proper authentication
	api.Use(func(c *fiber.Ctx) error {
		user := c.Get("X-User")
		logger.Debug("X-User header", zap.String("user", user))

		dbUser := infra.GetUserByName(user)
		if dbUser == nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Unknown user")
		}

		c.Locals("user", *dbUser)

		return c.Next()
	})

	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	// get workspace by name
	api.Get("/workspaces/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		user := c.Locals("user").(domain.User)

		canAccess, err := ketoClient.CheckUserAccessForWorkspaceName(name, user)
		if err != nil {
			logger.Error("Error checking user access with keto", zap.Error(err))
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if !canAccess {
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}

		workspace := infra.GetWorkspaceByName(name)
		if workspace == nil {
			return c.Status(404).SendString("Workspace not found")
		}

		return c.JSON(workspace)
	})

	// get all workspaces
	api.Get("/workspaces", func(c *fiber.Ctx) error {
		user := c.Locals("user").(domain.User)

		allowedWorkspaces, err := ketoClient.ListAccessibleWorkspaces(user)
		if err != nil {
			logger.Error("Error getting workspaces from keto", zap.Error(err))
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		// filter workspaces
		workspaces := make([]dto.WorkspaceAccess, len(allowedWorkspaces))
		for i := range allowedWorkspaces {
			w := infra.GetWorkspaceByName(allowedWorkspaces[i].WorkspaceName)

			if w == nil {
				logger.Error("Error getting workspace from db")
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			workspaces[i] = dto.WorkspaceAccess{
				WorkspaceName: w.Name,
				Owner:         allowedWorkspaces[i].Owner,
				User:          allowedWorkspaces[i].User,
			}
		}

		return c.JSON(workspaces)
	})

	// grant access to unit of a workspace if user is owner
	api.Post("/workspaces/:workspaceName/unit/:unitName/grant", func(c *fiber.Ctx) error {
		workspaceName := c.Params("workspaceName")
		unitName := c.Params("unitName")

		// TODO: check if user is owner of workspace
		workspace := infra.GetWorkspaceByName(workspaceName)
		if workspace == nil {
			return c.Status(404).SendString("Workspace not found")
		}

		unitExists := false
		for i := range workspace.Units {
			if workspace.Units[i].Name == unitName {
				unitExists = true
				break
			}
		}

		if !unitExists {
			return c.Status(404).SendString("Unit not found")
		}

		// grant permissions

		// return success
		return c.SendString("OK")
	})

}
