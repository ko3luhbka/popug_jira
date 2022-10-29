# Popug Jira schema registry

This package provides a kind of schema registry for storing & validating Kafka event schema versions. Only JSON Schema and Golang language are supported at the moment

## How to use

1. Create a new directory named after your event, e.g. "user". This stands for "schema type".
2. Create a JSON file with name containing only intergers, e.g. "1". This is a schema version.
3. Define your JSON Schema in the schema version file. A good place to get started with JSON Schema: https://json-schema.org/learn/getting-started-step-by-step.html
4. git commit & git push.
5. Import ``validator`` package in your Go code:
`import "github.com/ko3luhbka/schema_registry/validator"`
6. Validate your data:
````
var data = User{
	Name: "UserCreated",
	Version: 1,
	Data: UserInfo{
		ID: "51c0f139-d529-4035-9436-04b9fa0ad026",
        Name: "John"
	},
}

if err := validator.ValidateSchema(data, "user", 1); err != nil {
    return fmt.Errorf("invalid schema: %w", err)
} else {
    fmt.Println("schema is valid")
    return nil
}
````