package domain

// UserQuery is the documented bounded GET /users query surface.
type UserQuery struct {
	Page                                   PageRequest
	Query, Type, AccessTypePermissions     string
	IDs                                    []int64
	IncludeInactive                        *bool
	ExcludeDirectlyAddedMembersByEntityUID string
}
