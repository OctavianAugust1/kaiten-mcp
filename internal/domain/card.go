package domain

// CardQuery contains the documented non-beta GET /cards query surface. HTTP
// serialization belongs to the Kaiten adapter.
type CardQuery struct {
	CreatedBefore, CreatedAfter                           string
	UpdatedBefore, UpdatedAfter                           string
	FirstMovedInProgressAfter, FirstMovedInProgressBefore string
	LastMovedToDoneAfter, LastMovedToDoneBefore           string
	DueDateAfter, DueDateBefore                           string
	Query, Tag, ExternalID                                string
	Version, Limit, Offset                                int
	StartPosition                                         string
	TagIDs, TypeIDs                                       []int64
	ExcludeBoardIDs, ExcludeLaneIDs, ExcludeColumnIDs     []int64
	ColumnIDs, MemberIDs, OwnerIDs, ResponsibleIDs        []int64
	ExcludeOwnerIDs, ExcludeCardIDs, OrganizationIDs      []int64
	States                                                []int
	AdditionalCardFields, SearchFields                    []string
	OrderBy, OrderDirection                               []string
	SpaceID, OrderSpaceID, BoardID, ColumnID, LaneID      int64
	TypeID, ResponsibleID, OwnerID                        int64
	Condition                                             int
	IncludeSearchPreview, Archived, ASAP                  *bool
	Overdue, DoneOnTime, WithDueDate, IsRequest           *bool
	BrokenAPI                                             *bool
}
