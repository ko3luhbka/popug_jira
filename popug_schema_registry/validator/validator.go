package validator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/ko3luhbka/popug_schema_registry/schemas"
)

const schemaPathTemplate = "file://versions/%s/%d.json"

func Validate(data any, schemaType string, schemaVerion int) error {
	schemaPath := fmt.Sprintf(schemaPathTemplate, schemaType, schemaVerion)
	schemaLoader := gojsonschema.NewReferenceLoaderFileSystem(schemaPath, http.FS(schemas.SchemaFS))
	dataLoader := gojsonschema.NewGoLoader(data)
	res, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %v", err)
	}

	if res.Valid() {
		return nil
	}

	errors := make([]string, len(res.Errors()))
	for i, err := range res.Errors() {
		errors[i] = err.String()
	}

	return fmt.Errorf("invalid json: %v", strings.Join(errors, "; "))
}
