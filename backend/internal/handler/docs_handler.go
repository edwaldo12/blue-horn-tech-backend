package handler

import (
	"net/http"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/docs"
	"github.com/gin-gonic/gin"
)

// DocsHandler serves the embedded Swagger UI and OpenAPI spec.
type DocsHandler struct{}

// NewDocsHandler instantiates a documentation handler.
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// SwaggerUI serves the HTML page that boots Swagger UI.
func (h *DocsHandler) SwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", docs.SwaggerIndex())
}

// OpenAPIYaml serves the raw OpenAPI document.
func (h *DocsHandler) OpenAPIYaml(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml", docs.OpenAPIYAML())
}
