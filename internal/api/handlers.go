package api

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/konstantin/dmn-engine-go/internal/dmn"
	"github.com/konstantin/dmn-engine-go/internal/storage"
)

// Handler contains all HTTP handlers
type Handler struct {
	repo   storage.DefinitionRepository
	engine EngineInterface
	logger *slog.Logger
}

// EngineInterface is the interface for the evaluation engine
type EngineInterface interface {
	Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResult, error)
}

// EvaluateRequest mirrors engine.EvaluateRequest for API
type EvaluateRequest struct {
	DecisionKey string                 `json:"decisionKey"`
	Version     *int                   `json:"version,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
	TenantID    string                 `json:"tenantId,omitempty"`
}

// EvaluateResult mirrors engine.EvaluateResult for API
type EvaluateResult struct {
	DecisionKey  string                   `json:"decisionKey"`
	DecisionName string                   `json:"decisionName"`
	Version      int                      `json:"version"`
	Outputs      []map[string]interface{} `json:"outputs"`
	MatchedRules []string                 `json:"matchedRules"`
	EvaluatedAt  time.Time                `json:"evaluatedAt"`
	DurationNs   int64                    `json:"durationNs"`
}

// NewHandler creates a new handler
func NewHandler(repo storage.DefinitionRepository, engine EngineInterface, logger *slog.Logger) *Handler {
	return &Handler{
		repo:   repo,
		engine: engine,
		logger: logger,
	}
}

// ErrorResponse is an error response
type ErrorResponse struct {
	Error   string                `json:"error"`
	Details []dmn.ValidationError `json:"details,omitempty"`
}

// DeployRequest is a deploy request (for JSON body)
type DeployRequest struct {
	Name string `json:"name,omitempty"`
	XML  string `json:"xml"`
}

// DefinitionResponse is a simplified response without parsed model
type DefinitionResponse struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Version   int    `json:"version"`
	Name      string `json:"name"`
	Checksum  string `json:"checksum"`
	TenantID  string `json:"tenantId,omitempty"`
	CreatedAt string `json:"createdAt"`
}

func toDefinitionResponse(def *storage.Definition) *DefinitionResponse {
	return &DefinitionResponse{
		ID:        def.ID,
		Key:       def.Key,
		Version:   def.Version,
		Name:      def.Name,
		Checksum:  def.Checksum,
		TenantID:  def.TenantID,
		CreatedAt: def.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toDefinitionResponses(defs []*storage.Definition) []*DefinitionResponse {
	result := make([]*DefinitionResponse, len(defs))
	for i, def := range defs {
		result[i] = toDefinitionResponse(def)
	}
	return result
}

// DeployDefinition handles POST /api/v1/definitions
func (h *Handler) DeployDefinition(c *fiber.Ctx) error {
	var xmlContent []byte
	var name string

	// Try to get file from multipart form
	file, err := c.FormFile("file")
	if err == nil {
		f, err := file.Open()
		if err != nil {
			return c.Status(400).JSON(ErrorResponse{Error: "failed to open file"})
		}
		defer f.Close()
		xmlContent, _ = io.ReadAll(f)
		name = c.FormValue("name", file.Filename)
	} else {
		// Try JSON body
		contentType := c.Get("Content-Type")
		if contentType == "application/json" {
			var req DeployRequest
			if err := c.BodyParser(&req); err != nil {
				return c.Status(400).JSON(ErrorResponse{Error: "invalid JSON body"})
			}
			xmlContent = []byte(req.XML)
			name = req.Name
		} else {
			// Assume raw XML
			xmlContent = c.Body()
		}
	}

	if len(xmlContent) == 0 {
		return c.Status(400).JSON(ErrorResponse{Error: "no DMN content provided"})
	}

	// Parse DMN
	parser := dmn.NewParser()
	defs, err := parser.Parse(bytes.NewReader(xmlContent))
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "invalid DMN XML: " + err.Error()})
	}

	// Validate
	validator := dmn.NewValidator()
	if errors := validator.Validate(defs); len(errors) > 0 {
		return c.Status(400).JSON(ErrorResponse{
			Error:   "DMN validation failed",
			Details: errors,
		})
	}

	// Determine key and name
	key := ""
	if len(defs.Decisions) > 0 {
		key = defs.Decisions[0].ID
	} else {
		key = defs.ID
	}
	if name == "" {
		name = defs.Name
	}
	if name == "" {
		name = key
	}

	// Get tenant ID from header
	tenantID := c.Get("X-Tenant-ID")

	// Create definition
	def := &storage.Definition{
		Key:         key,
		Name:        name,
		Source:      string(xmlContent),
		ParsedModel: defs,
		TenantID:    tenantID,
	}

	// Deploy
	if err := h.repo.Deploy(c.Context(), def); err != nil {
		h.logger.Error("failed to deploy definition", "error", err)
		return c.Status(500).JSON(ErrorResponse{Error: "failed to deploy definition: " + err.Error()})
	}

	h.logger.Info("definition deployed",
		"key", def.Key,
		"version", def.Version,
		"tenantId", def.TenantID,
	)

	return c.Status(201).JSON(toDefinitionResponse(def))
}

// ListDefinitions handles GET /api/v1/definitions
func (h *Handler) ListDefinitions(c *fiber.Ctx) error {
	filter := &storage.ListFilter{
		Key:      c.Query("key"),
		TenantID: c.Get("X-Tenant-ID"),
		Limit:    c.QueryInt("limit", 100),
		Offset:   c.QueryInt("offset", 0),
	}

	defs, err := h.repo.List(c.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list definitions", "error", err)
		return c.Status(500).JSON(ErrorResponse{Error: "failed to list definitions"})
	}

	return c.JSON(toDefinitionResponses(defs))
}

// GetDefinition handles GET /api/v1/definitions/:key
func (h *Handler) GetDefinition(c *fiber.Ctx) error {
	key := c.Params("key")
	tenantID := c.Get("X-Tenant-ID")
	version := c.QueryInt("version", 0)

	var def *storage.Definition
	var err error

	if version > 0 {
		def, err = h.repo.GetByKeyAndVersion(c.Context(), key, version, tenantID)
	} else {
		def, err = h.repo.GetByKey(c.Context(), key, tenantID)
	}

	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "definition not found"})
	}

	return c.JSON(toDefinitionResponse(def))
}

// GetDefinitionXML handles GET /api/v1/definitions/:key/xml
func (h *Handler) GetDefinitionXML(c *fiber.Ctx) error {
	key := c.Params("key")
	tenantID := c.Get("X-Tenant-ID")
	version := c.QueryInt("version", 0)

	var def *storage.Definition
	var err error

	if version > 0 {
		def, err = h.repo.GetByKeyAndVersion(c.Context(), key, version, tenantID)
	} else {
		def, err = h.repo.GetByKey(c.Context(), key, tenantID)
	}

	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "definition not found"})
	}

	c.Set("Content-Type", "application/xml")
	c.Set("Content-Disposition", "attachment; filename="+def.Key+".dmn")
	return c.SendString(def.Source)
}

// GetDefinitionParsed handles GET /api/v1/definitions/:key/parsed
func (h *Handler) GetDefinitionParsed(c *fiber.Ctx) error {
	key := c.Params("key")
	tenantID := c.Get("X-Tenant-ID")
	version := c.QueryInt("version", 0)

	var def *storage.Definition
	var err error

	if version > 0 {
		def, err = h.repo.GetByKeyAndVersion(c.Context(), key, version, tenantID)
	} else {
		def, err = h.repo.GetByKey(c.Context(), key, tenantID)
	}

	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "definition not found"})
	}

	return c.JSON(def.ParsedModel)
}

// GetDefinitionVersions handles GET /api/v1/definitions/:key/versions
func (h *Handler) GetDefinitionVersions(c *fiber.Ctx) error {
	key := c.Params("key")
	tenantID := c.Get("X-Tenant-ID")

	// Используем List с фильтром по key
	filter := &storage.ListFilter{
		Key:      key,
		TenantID: tenantID,
	}

	defs, err := h.repo.List(c.Context(), filter)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "failed to get versions"})
	}

	if len(defs) == 0 {
		return c.Status(404).JSON(ErrorResponse{Error: "definition not found"})
	}

	// Возвращаем список версий
	type VersionInfo struct {
		Version   int    `json:"version"`
		CreatedAt string `json:"createdAt"`
		Checksum  string `json:"checksum"`
	}

	versions := make([]VersionInfo, len(defs))
	for i, def := range defs {
		versions[i] = VersionInfo{
			Version:   def.Version,
			CreatedAt: def.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Checksum:  def.Checksum,
		}
	}

	return c.JSON(fiber.Map{
		"key":      key,
		"versions": versions,
	})
}

// DeleteDefinition handles DELETE /api/v1/definitions/:key
func (h *Handler) DeleteDefinition(c *fiber.Ctx) error {
	key := c.Params("key")
	tenantID := c.Get("X-Tenant-ID")

	if err := h.repo.Delete(c.Context(), key, tenantID); err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "definition not found"})
	}

	h.logger.Info("definition deleted", "key", key, "tenantId", tenantID)
	return c.SendStatus(204)
}

// Health handles GET /health
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// Ready handles GET /ready
func (h *Handler) Ready(c *fiber.Ctx) error {
	// TODO: проверить подключение к БД
	return c.JSON(fiber.Map{
		"status": "ready",
	})
}

// Info handles GET /api/v1/info
func (h *Handler) Info(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":    "DMN Engine Go",
		"version": "0.1.0-pre-mvp",
		"features": fiber.Map{
			"dmn_version":   "1.3",
			"feel_support":  "basic",
			"storage":       "postgresql",
			"multi_tenancy": true,
			"hit_policies":  []string{"UNIQUE", "FIRST", "ANY", "PRIORITY", "COLLECT", "RULE ORDER", "OUTPUT ORDER"},
			"evaluation":    true, // Базовое выполнение реализовано
		},
	})
}

// Evaluate handles POST /api/v1/evaluate
func (h *Handler) Evaluate(c *fiber.Ctx) error {
	var req EvaluateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "invalid request body: " + err.Error()})
	}

	// Validate request
	if req.DecisionKey == "" {
		return c.Status(400).JSON(ErrorResponse{Error: "decisionKey is required"})
	}
	if req.Variables == nil {
		req.Variables = make(map[string]interface{})
	}

	// Get tenant ID from header if not in body
	if req.TenantID == "" {
		req.TenantID = c.Get("X-Tenant-ID")
	}

	// Evaluate
	if h.engine == nil {
		return c.Status(503).JSON(ErrorResponse{Error: "evaluation engine not available"})
	}

	result, err := h.engine.Evaluate(c.Context(), &req)
	if err != nil {
		h.logger.Error("evaluation failed",
			"decisionKey", req.DecisionKey,
			"error", err,
		)
		return c.Status(500).JSON(ErrorResponse{Error: "evaluation failed: " + err.Error()})
	}

	h.logger.Info("decision evaluated",
		"decisionKey", result.DecisionKey,
		"version", result.Version,
		"matchedRules", len(result.MatchedRules),
		"durationMs", result.DurationNs/1000000,
	)

	return c.JSON(result)
}
