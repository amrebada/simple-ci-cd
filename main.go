package main

import (
	"github.com/amrebada/deployment-server/core"
	"github.com/gofiber/fiber/v2"
	env "github.com/joho/godotenv"
)

func main() {
	// load Env variables
	env.Load()

	app := fiber.New()

	app.Get("/clean", func(c *fiber.Ctx) error {
		core.CleanDocker()
		return c.Status(200).JSON(map[string]string{"message": "Cleaned"})
	})
	app.Get("/:appId", core.HandleDeploymentRequest)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":3000")

}
