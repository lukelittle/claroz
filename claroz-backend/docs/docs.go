package docs

import "embed"

//go:embed swagger.json
var swaggerJSON embed.FS

// GetSwaggerJSON returns the embedded swagger.json file
func GetSwaggerJSON() []byte {
	data, err := swaggerJSON.ReadFile("swagger.json")
	if err != nil {
		panic(err)
	}
	return data
}
