package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	validID := "12345678-1234-1234-1234-123456789012"

	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
		wantErrs    []error
	}{
		{
			name: "valid user",
			in: User{
				ID: validID, Age: 25, Email: "alice@example.com",
				Role: "admin", Phones: []string{"79991234567"},
			},
		},
		{
			name:     "invalid user: multiple errors accumulate",
			in:       User{ID: "short", Age: 5, Email: "bad", Role: "unknown", Phones: []string{"1"}},
			wantErrs: []error{ErrStringLen, ErrIntMin, ErrStringRegexp, ErrStringIn, ErrStringLen},
		},
		{
			name: "valid app",
			in:   App{Version: "1.0.0"},
		},
		{
			name:     "invalid app: wrong version length",
			in:       App{Version: "v1"},
			wantErrs: []error{ErrStringLen},
		},
		{
			name: "token: no validate tags, passes",
			in:   Token{Header: []byte("h"), Payload: []byte("p"), Signature: []byte("s")},
		},
		{
			name: "valid response",
			in:   Response{Code: 200, Body: "OK"},
		},
		{
			name:     "invalid response: code not in set",
			in:       Response{Code: 301},
			wantErrs: []error{ErrIntIn},
		},
		{
			name:        "non-struct input returns ErrNotStruct",
			in:          "not a struct",
			expectedErr: ErrNotStruct,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d: %s", i, tt.name), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			if tt.wantErrs != nil {
				var ve ValidationErrors
				require.ErrorAs(t, err, &ve)
				require.Len(t, ve, len(tt.wantErrs))
				for j, wantErr := range tt.wantErrs {
					require.ErrorIs(t, ve[j].Err, wantErr)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestValidate_InvalidTags(t *testing.T) {
	t.Run("unknown rule", func(t *testing.T) {
		type S struct {
			F string `validate:"unknown:val"`
		}
		require.ErrorIs(t, Validate(S{F: "x"}), ErrInvalidTag)
	})

	t.Run("bad len param", func(t *testing.T) {
		type S struct {
			F string `validate:"len:abc"`
		}
		require.ErrorIs(t, Validate(S{F: "x"}), ErrInvalidTag)
	})

	t.Run("bad regexp", func(t *testing.T) {
		type S struct {
			F string `validate:"regexp:[invalid"`
		}
		require.ErrorIs(t, Validate(S{F: "x"}), ErrInvalidTag)
	})
}

func TestValidationErrors_Error(t *testing.T) {
	ve := ValidationErrors{
		{Field: "Name", Err: errors.New("too short")},
		{Field: "Age", Err: errors.New("out of range")},
	}
	s := ve.Error()
	require.Contains(t, s, "Name")
	require.Contains(t, s, "Age")
}
