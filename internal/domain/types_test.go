package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewIDRequiresPositiveValue(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		value   int64
		want    ID
		wantErr bool
	}{
		{name: "positive", value: 42, want: ID(42)},
		{name: "zero", value: 0, wantErr: true},
		{name: "negative", value: -1, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewID(%d) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewID(%d) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseUIDValidatesAndNormalizesCanonicalUUID(t *testing.T) {
	t.Parallel()

	const canonical = "550e8400-e29b-41d4-a716-446655440000"
	got, err := ParseUID(strings.ToUpper(canonical))
	if err != nil {
		t.Fatalf("ParseUID() error = %v", err)
	}
	if got.String() != canonical {
		t.Errorf("ParseUID() = %q, want %q", got, canonical)
	}

	for _, invalid := range []string{"", "550e8400e29b41d4a716446655440000", "550e8400-e29b-41d4-a716-44665544000z"} {
		if _, err := ParseUID(invalid); err == nil {
			t.Errorf("ParseUID(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestNewPageAppliesDefaultAndBounds(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name         string
		limit        int
		offset       int
		defaultLimit int
		want         PageRequest
		wantErr      bool
	}{
		{name: "default limit", defaultLimit: 25, want: PageRequest{Limit: 25}},
		{name: "maximum limit", limit: 100, offset: 4, defaultLimit: 25, want: PageRequest{Limit: 100, Offset: 4}},
		{name: "negative offset", limit: 10, offset: -1, defaultLimit: 25, wantErr: true},
		{name: "oversized limit", limit: 101, defaultLimit: 25, wantErr: true},
		{name: "invalid default", defaultLimit: 0, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewPage(tt.limit, tt.offset, tt.defaultLimit)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPage() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNewSearchPageEnforcesVersionSpecificContinuation(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		version       int
		offset        int
		startPosition string
		wantVersion   SearchVersion
		wantErr       bool
	}{
		{name: "default version one", wantVersion: SearchVersion1},
		{name: "version one offset", version: 1, offset: 10, wantVersion: SearchVersion1},
		{name: "version one rejects cursor", version: 1, startPosition: "next", wantErr: true},
		{name: "version two cursor", version: 2, startPosition: "next", wantVersion: SearchVersion2},
		{name: "version two rejects offset", version: 2, offset: 10, wantErr: true},
		{name: "unknown version", version: 3, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewSearchPage(tt.version, 20, tt.offset, tt.startPosition, 50)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSearchPage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.Version != tt.wantVersion {
				t.Errorf("NewSearchPage() version = %d, want %d", got.Version, tt.wantVersion)
			}
		})
	}
}

func TestResultEncodesAvailablePaginationMetadata(t *testing.T) {
	t.Parallel()

	offset := 0
	nextOffset := 20
	result := Result[[]Object]{
		Data: []Object{{"id": float64(1)}},
		Pagination: &Pagination{
			Limit:        20,
			Offset:       &offset,
			NextOffset:   &nextOffset,
			NextPosition: "cursor",
		},
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := `{"data":[{"id":1}],"pagination":{"limit":20,"offset":0,"next_offset":20,"next_position":"cursor"}}`
	if string(encoded) != want {
		t.Errorf("Marshal() = %s, want %s", encoded, want)
	}
}
