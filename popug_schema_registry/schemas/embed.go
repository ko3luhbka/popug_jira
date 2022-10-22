package schemas

import (
	"embed"
)

//go:embed versions/task/*.json
var SchemaFS embed.FS
