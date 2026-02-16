package docs

import "embed"

// SwaggerSpec holds the embedded swagger.yaml file.
//
//go:embed swagger.yaml
var SwaggerSpec embed.FS
