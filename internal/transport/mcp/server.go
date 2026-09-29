package mcptransport

import (
	"context"
	"errors"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// WorkspaceService is the MCP boundary's consumer-defined workspace port.
type WorkspaceService interface {
	ListSpaces(context.Context, domain.PageRequest) (domain.Result[[]domain.Object], error)
	GetSpace(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListBoards(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	GetBoard(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListColumns(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	ListSubcolumns(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	ListLanes(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
}

// CardService is the MCP boundary's consumer-defined card port.
type CardService interface {
	ListCards(context.Context, domain.CardQuery) (domain.Result[[]domain.Object], error)
	GetCard(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListCardComments(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	ListCardMembers(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
}

type UserService interface {
	ListUsers(context.Context, domain.UserQuery) (domain.Result[[]domain.Object], error)
	GetCurrentUser(context.Context) (domain.Result[domain.Object], error)
	ListSpaceUsers(context.Context, domain.ID) (domain.Result[[]domain.Object], error)
	GetSpaceUser(context.Context, domain.ID, domain.ID) (domain.Result[domain.Object], error)
}

type CatalogService interface {
	ListCardTypes(context.Context) (domain.Result[[]domain.Object], error)
	GetCardType(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListTags(context.Context) (domain.Result[[]domain.Object], error)
}

type DocumentService interface {
	ListDocuments(context.Context, domain.DocumentQuery) (domain.Result[[]domain.Object], error)
	GetDocument(context.Context, domain.UID) (domain.Result[domain.Object], error)
	GetDocumentSchema(context.Context, domain.ID) (domain.Result[domain.Object], error)
	ListDocumentGroups(context.Context) (domain.Result[[]domain.Object], error)
	GetDocumentGroup(context.Context, domain.UID) (domain.Result[domain.Object], error)
}

// Services holds manually injected feature use cases.
type Services struct {
	Workspace WorkspaceService
	Card      CardService
	User      UserService
	Catalog   CatalogService
	Document  DocumentService
}

// NewServer constructs a tools-only MCP server. Feature groups are registered
// only when their explicitly injected service is present.
func NewServer(version string, defaultPageLimit int, services Services) (*sdk.Server, error) {
	if version == "" {
		return nil, errors.New("server version is required")
	}
	if _, err := domain.NewPage(0, 0, defaultPageLimit); err != nil {
		return nil, err
	}
	server := sdk.NewServer(&sdk.Implementation{
		Name:        "kaiten-mcp",
		Title:       "Kaiten MCP",
		Description: "Read-only access to Kaiten",
		Version:     version,
	}, nil)
	if services.Workspace != nil {
		registerWorkspace(server, services.Workspace, defaultPageLimit)
	}
	if services.Card != nil {
		registerCard(server, services.Card)
	}
	if services.User != nil {
		registerUser(server, services.User, defaultPageLimit)
	}
	if services.Catalog != nil {
		registerCatalog(server, services.Catalog)
	}
	if services.Document != nil {
		registerDocument(server, services.Document)
	}
	return server, nil
}

type listSpacesInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"Maximum spaces to return, from 1 to 100"`
	Offset int `json:"offset,omitempty" jsonschema:"Number of spaces to skip; must not be negative"`
}

type spaceIDInput struct {
	SpaceID int64 `json:"space_id" jsonschema:"Positive numeric space identifier"`
}

type boardIDInput struct {
	BoardID int64 `json:"board_id" jsonschema:"Positive numeric board identifier"`
}

type columnIDInput struct {
	ColumnID int64 `json:"column_id" jsonschema:"Positive numeric column identifier"`
}

type cardIDInput struct {
	CardID int64 `json:"card_id" jsonschema:"Positive numeric card identifier"`
}

type userListInput struct {
	Limit                                  int     `json:"limit,omitempty"`
	Offset                                 int     `json:"offset,omitempty"`
	Query                                  string  `json:"query,omitempty"`
	Type                                   string  `json:"type,omitempty"`
	AccessTypePermissions                  string  `json:"access_type_permissions,omitempty"`
	IDs                                    []int64 `json:"ids,omitempty"`
	IncludeInactive                        *bool   `json:"include_inactive,omitempty"`
	ExcludeDirectlyAddedMembersByEntityUID string  `json:"exclude_directly_added_members_by_entity_uid,omitempty"`
}

type spaceUserDetailInput struct {
	SpaceID int64 `json:"space_id" jsonschema:"Positive numeric space identifier"`
	UserID  int64 `json:"user_id" jsonschema:"Positive numeric user identifier"`
}

type typeIDInput struct {
	TypeID int64 `json:"type_id"`
}

type documentListInput struct {
	Query                string   `json:"query,omitempty"`
	Condition            string   `json:"condition,omitempty"`
	Fields               []string `json:"fields,omitempty"`
	Version              int      `json:"version,omitempty"`
	Limit                int      `json:"limit,omitempty"`
	Offset               int      `json:"offset,omitempty"`
	StartPosition        string   `json:"start_position,omitempty"`
	IncludeSearchPreview *bool    `json:"include_search_preview,omitempty"`
}
type documentUIDInput struct {
	DocumentUID string `json:"document_uid"`
}
type documentSchemaInput struct {
	ID int64 `json:"id"`
}
type documentGroupUIDInput struct {
	DocumentGroupUID string `json:"document_group_uid"`
}

type cardListInput struct {
	CreatedBefore              string   `json:"created_before,omitempty"`
	CreatedAfter               string   `json:"created_after,omitempty"`
	UpdatedBefore              string   `json:"updated_before,omitempty"`
	UpdatedAfter               string   `json:"updated_after,omitempty"`
	FirstMovedInProgressAfter  string   `json:"first_moved_in_progress_after,omitempty"`
	FirstMovedInProgressBefore string   `json:"first_moved_in_progress_before,omitempty"`
	LastMovedToDoneAfter       string   `json:"last_moved_to_done_at_after,omitempty"`
	LastMovedToDoneBefore      string   `json:"last_moved_to_done_at_before,omitempty"`
	DueDateAfter               string   `json:"due_date_after,omitempty"`
	DueDateBefore              string   `json:"due_date_before,omitempty"`
	Query                      string   `json:"query,omitempty"`
	Version                    int      `json:"version,omitempty"`
	Tag                        string   `json:"tag,omitempty"`
	TagIDs                     []int64  `json:"tag_ids,omitempty"`
	TypeIDs                    []int64  `json:"type_ids,omitempty"`
	ExcludeBoardIDs            []int64  `json:"exclude_board_ids,omitempty"`
	ExcludeLaneIDs             []int64  `json:"exclude_lane_ids,omitempty"`
	ExcludeColumnIDs           []int64  `json:"exclude_column_ids,omitempty"`
	ColumnIDs                  []int64  `json:"column_ids,omitempty"`
	MemberIDs                  []int64  `json:"member_ids,omitempty"`
	OwnerIDs                   []int64  `json:"owner_ids,omitempty"`
	ResponsibleIDs             []int64  `json:"responsible_ids,omitempty"`
	States                     []int    `json:"states,omitempty"`
	ExternalID                 string   `json:"external_id,omitempty"`
	AdditionalCardFields       []string `json:"additional_card_fields,omitempty"`
	SearchFields               []string `json:"search_fields,omitempty"`
	SpaceID                    int64    `json:"space_id,omitempty"`
	Limit                      int      `json:"limit,omitempty"`
	Offset                     int      `json:"offset,omitempty"`
	StartPosition              string   `json:"start_position,omitempty"`
	IncludeSearchPreview       *bool    `json:"include_search_preview,omitempty"`
	OrderSpaceID               int64    `json:"order_space_id,omitempty"`
	BoardID                    int64    `json:"board_id,omitempty"`
	ColumnID                   int64    `json:"column_id,omitempty"`
	LaneID                     int64    `json:"lane_id,omitempty"`
	Condition                  int      `json:"condition,omitempty"`
	TypeID                     int64    `json:"type_id,omitempty"`
	ResponsibleID              int64    `json:"responsible_id,omitempty"`
	OwnerID                    int64    `json:"owner_id,omitempty"`
	Archived                   *bool    `json:"archived,omitempty"`
	ASAP                       *bool    `json:"asap,omitempty"`
	Overdue                    *bool    `json:"overdue,omitempty"`
	DoneOnTime                 *bool    `json:"done_on_time,omitempty"`
	WithDueDate                *bool    `json:"with_due_date,omitempty"`
	OrderBy                    []string `json:"order_by,omitempty"`
	OrderDirection             []string `json:"order_direction,omitempty"`
	IsRequest                  *bool    `json:"is_request,omitempty"`
	ExcludeOwnerIDs            []int64  `json:"exclude_owner_ids,omitempty"`
	ExcludeCardIDs             []int64  `json:"exclude_card_ids,omitempty"`
	OrganizationIDs            []int64  `json:"organizations_ids,omitempty"`
	BrokenAPI                  *bool    `json:"broken_api,omitempty"`
}

func registerWorkspace(server *sdk.Server, service WorkspaceService, defaultPageLimit int) {
	sdk.AddTool(server, readTool("kaiten_list_spaces", "List one bounded page of spaces"), func(
		ctx context.Context,
		_ *sdk.CallToolRequest,
		input listSpacesInput,
	) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		page, err := domain.NewPage(input.Limit, input.Offset, defaultPageLimit)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := service.ListSpaces(ctx, page)
		return nil, output, err
	})

	sdk.AddTool(server, readTool("kaiten_get_space", "Get a space by ID"), func(
		ctx context.Context,
		_ *sdk.CallToolRequest,
		input spaceIDInput,
	) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		id, err := domain.NewID(input.SpaceID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetSpace(ctx, id)
		return nil, output, err
	})

	sdk.AddTool(server, readTool("kaiten_list_boards", "List boards in a space"), listBySpaceHandler(service.ListBoards))
	sdk.AddTool(server, readTool("kaiten_get_board", "Get a board by ID"), func(
		ctx context.Context,
		_ *sdk.CallToolRequest,
		input boardIDInput,
	) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		id, err := domain.NewID(input.BoardID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetBoard(ctx, id)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_list_columns", "List columns in a board"), listByBoardHandler(service.ListColumns))
	sdk.AddTool(server, readTool("kaiten_list_subcolumns", "List subcolumns in a column"), func(
		ctx context.Context,
		_ *sdk.CallToolRequest,
		input columnIDInput,
	) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		id, err := domain.NewID(input.ColumnID)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := service.ListSubcolumns(ctx, id)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_list_lanes", "List lanes in a board"), listByBoardHandler(service.ListLanes))
}

func registerCard(server *sdk.Server, service CardService) {
	sdk.AddTool(server, readTool("kaiten_list_cards", "List one bounded page of cards using documented non-beta filters"), func(
		ctx context.Context,
		_ *sdk.CallToolRequest,
		input cardListInput,
	) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		output, err := service.ListCards(ctx, input.toDomain())
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_card", "Get a card by ID"), func(
		ctx context.Context,
		_ *sdk.CallToolRequest,
		input cardIDInput,
	) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		id, err := domain.NewID(input.CardID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetCard(ctx, id)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_list_card_comments", "List comments on a card"), listByCardHandler(service.ListCardComments))
	sdk.AddTool(server, readTool("kaiten_list_card_members", "List members of a card"), listByCardHandler(service.ListCardMembers))
}

func registerUser(server *sdk.Server, service UserService, defaultPageLimit int) {
	sdk.AddTool(server, readTool("kaiten_list_users", "List one bounded page of users"), func(ctx context.Context, _ *sdk.CallToolRequest, input userListInput) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		page, err := domain.NewPage(input.Limit, input.Offset, defaultPageLimit)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := service.ListUsers(ctx, domain.UserQuery{Page: page, Query: input.Query, Type: input.Type, AccessTypePermissions: input.AccessTypePermissions, IDs: input.IDs, IncludeInactive: input.IncludeInactive, ExcludeDirectlyAddedMembersByEntityUID: input.ExcludeDirectlyAddedMembersByEntityUID})
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_current_user", "Get the user for the configured token"), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		output, err := service.GetCurrentUser(ctx)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_list_space_users", "List users in a space"), func(ctx context.Context, _ *sdk.CallToolRequest, input spaceIDInput) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		spaceID, err := domain.NewID(input.SpaceID)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := service.ListSpaceUsers(ctx, spaceID)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_space_user", "Get a user in a space"), func(ctx context.Context, _ *sdk.CallToolRequest, input spaceUserDetailInput) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		spaceID, err := domain.NewID(input.SpaceID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		userID, err := domain.NewID(input.UserID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetSpaceUser(ctx, spaceID, userID)
		return nil, output, err
	})
}

func registerCatalog(server *sdk.Server, service CatalogService) {
	sdk.AddTool(server, readTool("kaiten_list_card_types", "List card types"), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		output, err := service.ListCardTypes(ctx)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_card_type", "Get a card type by ID"), func(ctx context.Context, _ *sdk.CallToolRequest, input typeIDInput) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		id, err := domain.NewID(input.TypeID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetCardType(ctx, id)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_list_tags", "List tags"), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		output, err := service.ListTags(ctx)
		return nil, output, err
	})
}

func registerDocument(server *sdk.Server, service DocumentService) {
	sdk.AddTool(server, readTool("kaiten_list_documents", "List one bounded page of documents"), func(ctx context.Context, _ *sdk.CallToolRequest, input documentListInput) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		output, err := service.ListDocuments(ctx, domain.DocumentQuery{Query: input.Query, Condition: input.Condition, Fields: input.Fields, Version: input.Version, Limit: input.Limit, Offset: input.Offset, StartPosition: input.StartPosition, IncludeSearchPreview: input.IncludeSearchPreview})
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_document", "Get a document by UUID"), func(ctx context.Context, _ *sdk.CallToolRequest, input documentUIDInput) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		uid, err := domain.ParseUID(input.DocumentUID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetDocument(ctx, uid)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_document_schema", "Get a document schema by ID"), func(ctx context.Context, _ *sdk.CallToolRequest, input documentSchemaInput) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		id, err := domain.NewID(input.ID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetDocumentSchema(ctx, id)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_list_document_groups", "List document groups"), func(ctx context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		output, err := service.ListDocumentGroups(ctx)
		return nil, output, err
	})
	sdk.AddTool(server, readTool("kaiten_get_document_group", "Get a document group by UUID"), func(ctx context.Context, _ *sdk.CallToolRequest, input documentGroupUIDInput) (*sdk.CallToolResult, domain.Result[domain.Object], error) {
		uid, err := domain.ParseUID(input.DocumentGroupUID)
		if err != nil {
			return nil, domain.Result[domain.Object]{}, err
		}
		output, err := service.GetDocumentGroup(ctx, uid)
		return nil, output, err
	})
}

func (input cardListInput) toDomain() domain.CardQuery {
	return domain.CardQuery{
		CreatedBefore: input.CreatedBefore, CreatedAfter: input.CreatedAfter, UpdatedBefore: input.UpdatedBefore, UpdatedAfter: input.UpdatedAfter,
		FirstMovedInProgressAfter: input.FirstMovedInProgressAfter, FirstMovedInProgressBefore: input.FirstMovedInProgressBefore,
		LastMovedToDoneAfter: input.LastMovedToDoneAfter, LastMovedToDoneBefore: input.LastMovedToDoneBefore, DueDateAfter: input.DueDateAfter, DueDateBefore: input.DueDateBefore,
		Query: input.Query, Version: input.Version, Tag: input.Tag, TagIDs: input.TagIDs, TypeIDs: input.TypeIDs,
		ExcludeBoardIDs: input.ExcludeBoardIDs, ExcludeLaneIDs: input.ExcludeLaneIDs, ExcludeColumnIDs: input.ExcludeColumnIDs,
		ColumnIDs: input.ColumnIDs, MemberIDs: input.MemberIDs, OwnerIDs: input.OwnerIDs, ResponsibleIDs: input.ResponsibleIDs,
		States: input.States, ExternalID: input.ExternalID, AdditionalCardFields: input.AdditionalCardFields, SearchFields: input.SearchFields,
		SpaceID: input.SpaceID, Limit: input.Limit, Offset: input.Offset, StartPosition: input.StartPosition, IncludeSearchPreview: input.IncludeSearchPreview,
		OrderSpaceID: input.OrderSpaceID, BoardID: input.BoardID, ColumnID: input.ColumnID, LaneID: input.LaneID, Condition: input.Condition,
		TypeID: input.TypeID, ResponsibleID: input.ResponsibleID, OwnerID: input.OwnerID, Archived: input.Archived, ASAP: input.ASAP,
		Overdue: input.Overdue, DoneOnTime: input.DoneOnTime, WithDueDate: input.WithDueDate, OrderBy: input.OrderBy, OrderDirection: input.OrderDirection,
		IsRequest: input.IsRequest, ExcludeOwnerIDs: input.ExcludeOwnerIDs, ExcludeCardIDs: input.ExcludeCardIDs, OrganizationIDs: input.OrganizationIDs, BrokenAPI: input.BrokenAPI,
	}
}

func listBySpaceHandler(call func(context.Context, domain.ID) (domain.Result[[]domain.Object], error)) sdk.ToolHandlerFor[spaceIDInput, domain.Result[[]domain.Object]] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, input spaceIDInput) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		id, err := domain.NewID(input.SpaceID)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := call(ctx, id)
		return nil, output, err
	}
}

func listByBoardHandler(call func(context.Context, domain.ID) (domain.Result[[]domain.Object], error)) sdk.ToolHandlerFor[boardIDInput, domain.Result[[]domain.Object]] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, input boardIDInput) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		id, err := domain.NewID(input.BoardID)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := call(ctx, id)
		return nil, output, err
	}
}

func listByCardHandler(call func(context.Context, domain.ID) (domain.Result[[]domain.Object], error)) sdk.ToolHandlerFor[cardIDInput, domain.Result[[]domain.Object]] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, input cardIDInput) (*sdk.CallToolResult, domain.Result[[]domain.Object], error) {
		id, err := domain.NewID(input.CardID)
		if err != nil {
			return nil, domain.Result[[]domain.Object]{}, err
		}
		output, err := call(ctx, id)
		return nil, output, err
	}
}

func readTool(name, description string) *sdk.Tool {
	destructive := false
	openWorld := true
	return &sdk.Tool{
		Name:        name,
		Description: description,
		Annotations: &sdk.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: &destructive,
			IdempotentHint:  true,
			OpenWorldHint:   &openWorld,
		},
	}
}
