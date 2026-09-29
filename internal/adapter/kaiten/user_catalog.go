package kaiten

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/OctavianAugust1/kaiten-mcp/internal/domain"
)

// ListUsers reads one bounded page of company users.
func (c *Client) ListUsers(ctx context.Context, query domain.UserQuery) (domain.Result[[]domain.Object], error) {
	values := url.Values{"limit": {strconv.Itoa(query.Page.Limit)}, "offset": {strconv.Itoa(query.Page.Offset)}}
	setString(values, "query", query.Query)
	setString(values, "type", query.Type)
	setString(values, "access_type_permissions", query.AccessTypePermissions)
	setString(values, "exclude_directly_added_members_by_entity_uid", query.ExcludeDirectlyAddedMembersByEntityUID)
	if query.IncludeInactive != nil {
		values.Set("include_inactive", strconv.FormatBool(*query.IncludeInactive))
	}
	if len(query.IDs) > 0 {
		ids, err := positiveIDs(query.IDs)
		if err != nil {
			return domain.Result[[]domain.Object]{}, &domain.Error{Kind: domain.ErrorInvalidArgument, Operation: "validate user query", Cause: errors.New("ids must be positive")}
		}
		values.Set("ids", ids)
	}
	return c.list(ctx, "users", values, &query.Page)
}

// GetCurrentUser reads the user represented by the configured token.
func (c *Client) GetCurrentUser(ctx context.Context) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "users/current")
}

// ListSpaceUsers reads users in a space.
func (c *Client) ListSpaceUsers(ctx context.Context, spaceID domain.ID) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "spaces/"+strconv.FormatInt(int64(spaceID), 10)+"/users", nil, nil)
}

// GetSpaceUser reads one space-user membership.
func (c *Client) GetSpaceUser(ctx context.Context, spaceID, userID domain.ID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "spaces/"+strconv.FormatInt(int64(spaceID), 10)+"/users/"+strconv.FormatInt(int64(userID), 10))
}

// ListCardTypes reads card types.
func (c *Client) ListCardTypes(ctx context.Context) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "card-types", nil, nil)
}

// GetCardType reads one card type.
func (c *Client) GetCardType(ctx context.Context, typeID domain.ID) (domain.Result[domain.Object], error) {
	return c.detail(ctx, "card-types/"+strconv.FormatInt(int64(typeID), 10))
}

// ListTags reads available tags.
func (c *Client) ListTags(ctx context.Context) (domain.Result[[]domain.Object], error) {
	return c.list(ctx, "tags", nil, nil)
}
