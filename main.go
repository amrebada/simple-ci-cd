package main

import (
	"github.com/amrebada/deployment-server/core"
	"github.com/gofiber/fiber/v2"
	env "github.com/joho/godotenv"
)

const PORT = ":7543"

func main() {
	// load Env variables
	env.Load()

	app := fiber.New()

	manage := app.Group("/manage", core.AuthenticateAPIKey)
	manage.Get("/clean", func(c *fiber.Ctx) error {
		core.CleanDocker()
		return c.Status(200).JSON(map[string]string{"message": "Cleaned"})
	})
	manage.Get("/:appId", core.HandleDeploymentRequest)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Deployment Server working")
	})

	app.ListenTLS(PORT, "cert.crt", "cert.key")

}
