package docs

import _ "embed"

//go:embed openapi.yaml
var openAPISpec []byte

//go:embed index.html
var swaggerHTML []byte

// OpenAPIYAML returns the embedded OpenAPI specification.
func OpenAPIYAML() []byte {
	return openAPISpec
}

// SwaggerIndex returns the embedded Swagger UI HTML shell.
func SwaggerIndex() []byte {
	return swaggerHTML
}
