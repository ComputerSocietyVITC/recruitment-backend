package types

type ExampleJSONBody struct {
	Name string `json:"name" binding:"required"`
}
