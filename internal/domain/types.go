package domain

import (
	"encoding/hex"
	"errors"
	"strings"
)

// ID is a validated positive Kaiten numeric identifier.
type ID int64

// NewID rejects zero and negative identifiers before a path is constructed.
func NewID(value int64) (ID, error) {
	if value <= 0 {
		return 0, errors.New("identifier must be positive")
	}
	return ID(value), nil
}

// UID is a normalized canonical UUID used by documents and document groups.
type UID string

// ParseUID validates the 8-4-4-4-12 UUID representation and normalizes hex
// letters to lowercase.
func ParseUID(value string) (UID, error) {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return "", errors.New("UID must be a canonical UUID")
	}
	hexValue := strings.NewReplacer("-", "").Replace(value)
	if _, err := hex.DecodeString(hexValue); err != nil {
		return "", errors.New("UID must be a canonical UUID")
	}
	return UID(strings.ToLower(value)), nil
}

// String returns the normalized UID representation.
func (u UID) String() string {
	return string(u)
}

// PageRequest is one validated offset-based request. Limit never exceeds 100.
type PageRequest struct {
	Limit  int
	Offset int
}

// NewPage applies the configured default when limit is zero and validates one
// bounded page request.
func NewPage(limit, offset, defaultLimit int) (PageRequest, error) {
	if defaultLimit < 1 || defaultLimit > 100 {
		return PageRequest{}, errors.New("default page limit must be between 1 and 100")
	}
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 1 || limit > 100 {
		return PageRequest{}, errors.New("page limit must be between 1 and 100")
	}
	if offset < 0 {
		return PageRequest{}, errors.New("page offset must not be negative")
	}
	return PageRequest{Limit: limit, Offset: offset}, nil
}

// SearchVersion selects Kaiten's offset or cursor search contract.
type SearchVersion int

const (
	// SearchVersion1 uses offset pagination.
	SearchVersion1 SearchVersion = 1
	// SearchVersion2 uses start_position cursor pagination.
	SearchVersion2 SearchVersion = 2
)

// SearchPage combines a bounded page with its version-specific continuation.
type SearchPage struct {
	PageRequest
	Version       SearchVersion
	StartPosition string
}

// NewSearchPage validates that offset and cursor continuation fields are not
// mixed across search protocol versions.
func NewSearchPage(version, limit, offset int, startPosition string, defaultLimit int) (SearchPage, error) {
	page, err := NewPage(limit, offset, defaultLimit)
	if err != nil {
		return SearchPage{}, err
	}
	if version == 0 {
		version = int(SearchVersion1)
	}
	switch SearchVersion(version) {
	case SearchVersion1:
		if startPosition != "" {
			return SearchPage{}, errors.New("start_position requires search version 2")
		}
	case SearchVersion2:
		if offset != 0 {
			return SearchPage{}, errors.New("offset requires search version 1")
		}
	default:
		return SearchPage{}, errors.New("search version must be 1 or 2")
	}
	return SearchPage{
		PageRequest:   page,
		Version:       SearchVersion(version),
		StartPosition: startPosition,
	}, nil
}

// Object preserves evolving Kaiten JSON objects without discarding fields.
type Object map[string]any

// Pagination reports only continuation metadata available for a list result.
type Pagination struct {
	Limit        int    `json:"limit"`
	Offset       *int   `json:"offset,omitempty"`
	NextOffset   *int   `json:"next_offset,omitempty"`
	NextPosition string `json:"next_position,omitempty"`
}

// Result is the stable tool envelope for both detail and list reads.
type Result[T any] struct {
	Data       T           `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}
