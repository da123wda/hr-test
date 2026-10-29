package model

import "time"

type SchemaStatus string

const (
	SchemaStatusDraft     SchemaStatus = "draft"
	SchemaStatusPublished SchemaStatus = "published"
)

type FormSchema struct {
	Id               string            `json:"id" xorm:"varchar(64) pk 'id'"`
	FormId           string            `json:"formId" xorm:"varchar(64) unique notnull 'form_id'"`
	Title            string            `json:"title" xorm:"varchar(255) notnull 'title'"`
	Description      string            `json:"description" xorm:"text 'description'"`
	Version          int               `json:"version" xorm:"int notnull default 1 'version'"`
	PublishedVersion int               `json:"publishedVersion" xorm:"int notnull default 0 'published_version'"`
	Status           SchemaStatus      `json:"status" xorm:"varchar(32) notnull default 'draft' 'status'"`
	Fields           []FieldDefinition `json:"fields" xorm:"jsonb notnull 'fields'"`
	Migrations       map[string]string `json:"migrations,omitempty" xorm:"jsonb 'migrations'"`
	CreatedAt        time.Time         `json:"createdAt" xorm:"created 'created_at'"`
	UpdatedAt        time.Time         `json:"updatedAt" xorm:"updated 'updated_at'"`
	CreatedBy        string            `json:"createdBy" xorm:"varchar(64) 'created_by'"`
}

type FieldDefinition struct {
	Key          string            `json:"key"`
	Label        string            `json:"label"`
	Type         string            `json:"type"`
	Placeholder  string            `json:"placeholder,omitempty"`
	DefaultValue any               `json:"defaultValue,omitempty"`
	Deprecated   bool              `json:"deprecated,omitempty"`
	Validation   *ValidationRules  `json:"validation,omitempty"`
	Options      []OptionItem      `json:"options,omitempty"`
	Children     []FieldDefinition `json:"children,omitempty"`
	ItemFields   []FieldDefinition `json:"itemFields,omitempty"`
	Rules        []FieldRule       `json:"rules,omitempty"`
	GridCol      int               `json:"gridCol,omitempty"`
	AlertConfig  *AlertConfig      `json:"alertConfig,omitempty"`
}

type ValidationRules struct {
	Required      bool     `json:"required,omitempty"`
	Email         bool     `json:"email,omitempty"`
	Regex         string   `json:"regex,omitempty"`
	Min           *float64 `json:"min,omitempty"`
	Max           *float64 `json:"max,omitempty"`
	MinLength     *int     `json:"minLength,omitempty"`
	MaxLength     *int     `json:"maxLength,omitempty"`
	CustomMessage string   `json:"customMessage,omitempty"`
}

type OptionItem struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type FieldRule struct {
	TargetKey  string         `json:"targetKey"`
	Action     string         `json:"action"`
	Expression string         `json:"expression"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type AlertConfig struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

type FormSchemaVersion struct {
	Id          int64             `json:"id" xorm:"pk autoincr 'id'"`
	FormId      string            `json:"formId" xorm:"varchar(64) notnull 'form_id'"`
	Version     int               `json:"version" xorm:"int notnull 'version'"`
	Title       string            `json:"title" xorm:"varchar(255) notnull 'title'"`
	Description string            `json:"description" xorm:"text 'description'"`
	Fields      []FieldDefinition `json:"fields" xorm:"jsonb notnull 'fields'"`
	PublishedAt time.Time         `json:"publishedAt" xorm:"created 'published_at'"`
	PublishedBy string            `json:"publishedBy" xorm:"varchar(64) 'published_by'"`
}
