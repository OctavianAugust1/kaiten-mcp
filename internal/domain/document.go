package domain

// DocumentQuery contains documented GET /documents search inputs.
type DocumentQuery struct {
	Query, Condition       string
	Fields                 []string
	Version, Limit, Offset int
	StartPosition          string
	IncludeSearchPreview   *bool
}
