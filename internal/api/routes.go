package api

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App, h *Handler) {
	// Health endpoints
	app.Get("/health", h.Health)
	app.Get("/ready", h.Ready)

	// API v1
	v1 := app.Group("/api/v1")

	// Info
	v1.Get("/info", h.Info)

	// Definitions (Decision Definition management)
	definitions := v1.Group("/definitions")
	definitions.Post("/", h.DeployDefinition)                  // Deploy new definition
	definitions.Get("/", h.ListDefinitions)                    // List all definitions
	definitions.Get("/:key", h.GetDefinition)                  // Get definition by key
	definitions.Get("/:key/xml", h.GetDefinitionXML)           // Get original XML
	definitions.Get("/:key/parsed", h.GetDefinitionParsed)     // Get parsed model
	definitions.Get("/:key/versions", h.GetDefinitionVersions) // Get all versions
	definitions.Delete("/:key", h.DeleteDefinition)            // Delete definition

	// Evaluation
	v1.Post("/evaluate", h.Evaluate) // Evaluate a decision
}
