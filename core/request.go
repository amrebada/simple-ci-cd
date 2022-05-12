package core

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthenticateAPIKey(c *fiber.Ctx) error {
	apiKey := c.GetReqHeaders()["X-Api-Key"]
	if apiKey == "" || apiKey != os.Getenv("API_KEY") {
		return errors.New("API key is not valid")
	}
	return nil

}

func HandleDeploymentRequest(c *fiber.Ctx) error {
	if err := AuthenticateAPIKey(c); err != nil {
		return c.Status(401).JSON(map[string]string{"error": err.Error()})
	}

	appId := strings.ToUpper(c.Params("appId"))
	if appId == "" {
		return c.Status(400).JSON(map[string]string{"error": "appId is required"})
	}

	repositoryPath := os.Getenv(fmt.Sprintf("REPOSITORY_PATH_%s", appId))
	repositoryUrl := os.Getenv(fmt.Sprintf("REPOSITORY_URL_%s", appId))
	if repositoryPath == "" || repositoryUrl == "" {
		return c.Status(400).JSON(map[string]string{"error": "appId is not valid"})
	}

	if err := RemoveRepository(repositoryPath); err != nil {
		return c.Status(500).JSON(map[string]string{"error": err.Error(), "errorType": "Cloning"})
	}

	if err := Clone(repositoryUrl, repositoryPath); err != nil {
		return c.Status(500).JSON(map[string]string{"error": err.Error(), "errorType": "Cloning"})
	}

	warnings := map[string]string{}

	if err := CopyDotEnvToDir(repositoryPath, appId); err != nil {
		warnings["dotenv"] = err.Error()
	}

	go BuildAndRunDockerContainer(repositoryPath, appId)

	return c.Status(200).JSON(map[string]interface{}{"message": "Deployment successful", "warnings": warnings})
}
