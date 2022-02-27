package main

type Swagger struct {
	Swagger     string               `yaml:"swagger" json:"swagger"`
	Info        *Info                `yaml:"info" json:"info"`
	Paths       map[string]*PathItem `yaml:"paths" json:"paths"`
	Definitions map[string]*Schema   `yaml:"definitions" json:"definitions"`
}

type Info struct {
	Version string `yaml:"version" json:"version"`
	Title   string `yaml:"title" json:"title"`
}

type PathItem struct {
	Get    *Operation `yaml:"get" json:"get"`
	Post   *Operation `yaml:"post" json:"post"`
	Put    *Operation `yaml:"put" json:"put"`
	Patch  *Operation `yaml:"patch" json:"patch"`
	Delete *Operation `yaml:"delete" json:"delete"`
}

type Operation struct {
	Tags        []string             `yaml:"tags" json:"tags"`
	Summary     string               `yaml:"summary" json:"summary"`
	Description string               `yaml:"description" json:"description"`
	OperationID string               `yaml:"operationId" json:"operationId"`
	Parameters  []*Parameter         `yaml:"parameters" json:"parameters"`
	Responses   map[string]*Response `yaml:"responses" json:"responses"`
	Security    []*Security          `yaml:"security" json:"security"`
}

type Parameter struct {
	Name        string `yaml:"name" json:"name"`
	In          string `yaml:"in" json:"in"`
	Description string `yaml:"description" json:"description,omitempty"`
	Required    bool   `yaml:"required" json:"required"`

	Schema *Schema `yaml:"schema" json:"schema"`

	Type    string      `yaml:"type" json:"type,omitempty"`
	Format  string      `yaml:"format,omitempty" json:"format,omitempty"`
	Items   *Schema     `yaml:"items,omitempty" json:"items,omitempty"`
	Default interface{} `yaml:"default,omitempty" json:"default,omitempty"`
	Maximum interface{} `yaml:"maximum,omitempty" json:"maximum,omitempty"`

	Examples interface{} `yaml:"x-example" json:"x-example"`
}

type Response struct {
	Description string                 `yaml:"description" json:"description"`
	Schema      *Schema                `yaml:"schema" json:"schema"`
	Examples    map[string]interface{} `yaml:"examples" json:"examples"`
}

type Security struct {
	Roles []string `yaml:"roles" json:"roles"`
}

type Schema struct {
	Ref                  string             `yaml:"$ref,omitempty" json:"$ref,omitempty"`
	Format               string             `yaml:"format,omitempty" json:"format,omitempty"`
	Title                string             `yaml:"title,omitempty" json:"title,omitempty"`
	Description          string             `yaml:"description" json:"description,omitempty"`
	Default              interface{}        `yaml:"default,omitempty" json:"default,omitempty"`
	Maximum              interface{}        `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	Items                *Schema            `yaml:"items,omitempty" json:"items,omitempty"`
	Type                 string             `yaml:"type" json:"type,omitempty"`
	AdditionalProperties *Schema            `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	Enum                 []string           `yaml:"enum,omitempty" json:"enum,omitempty"`
	Properties           map[string]*Schema `yaml:"properties" json:"properties,omitempty"`
	Required             []string           `yaml:"required" json:"required,omitempty"`
}
