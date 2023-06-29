package main

import (
	"github.com/ChristianSch/keto-go/api/infra"
	"github.com/ChristianSch/keto-go/api/infra/http"
	"github.com/ChristianSch/keto-go/api/infra/ory/keto"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	logger := infra.InitLogger()
	// replace global logger
	undo := zap.ReplaceGlobals(&logger)
	defer undo()

	// keto client
	ketoClient := keto.NewKetoClient(&logger)

	// using go fiber
	api := fiber.New()

	// register routes
	http.ApiFactory(api, &logger, ketoClient)

	// start server
	logger.Sugar().Fatal(api.Listen(":3210"))
}
