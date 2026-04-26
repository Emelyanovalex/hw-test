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
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
		wantErrs    []error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Alice",
				Age:    25,
				Email:  "alice@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			expectedErr: nil,
		},
		{
			name: "invalid user: wrong ID length",
			in: User{
				ID:     "short-id",
				Name:   "Bob",
				Age:    30,
				Email:  "bob@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			wantErrs: []error{ErrStringLen},
		},
		{
			name: "invalid user: age too low",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Teen",
				Age:    16,
				Email:  "teen@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			wantErrs: []error{ErrIntMin},
		},
		{
			name: "invalid user: age too high",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Old",
				Age:    60,
				Email:  "old@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			wantErrs: []error{ErrIntMax},
		},
		{
			name: "invalid user: bad email",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Eve",
				Age:    25,
				Email:  "not-an-email",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			wantErrs: []error{ErrStringRegexp},
		},
		{
			name: "invalid user: bad role",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Eve",
				Age:    25,
				Email:  "eve@example.com",
				Role:   "superuser",
				Phones: []string{"79991234567"},
			},
			wantErrs: []error{ErrStringIn},
		},
		{
			name: "invalid user: bad phone length in slice",
			in: User{
				ID:     "12345678-1234-1234-1234-123456789012",
				Name:   "Eve",
				Age:    25,
				Email:  "eve@example.com",
				Role:   "admin",
				Phones: []string{"123"},
			},
			wantErrs: []error{ErrStringLen},
		},
		{
			name: "valid app",
			in:   App{Version: "1.0.0"},
			expectedErr: nil,
		},
		{
			name:     "invalid app: version length",
			in:       App{Version: "v1"},
			wantErrs: []error{ErrStringLen},
		},
		{
			name:        "token: no validate tags, passes",
			in:          Token{Header: []byte("h"), Payload: []byte("p"), Signature: []byte("s")},
			expectedErr: nil,
		},
		{
			name:        "valid response",
			in:          Response{Code: 200, Body: "OK"},
			expectedErr: nil,
		},
		{
			name:        "valid response 404",
			in:          Response{Code: 404},
			expectedErr: nil,
		},
		{
			name:     "invalid response: code not in set",
			in:       Response{Code: 301, Body: "Moved"},
			wantErrs: []error{ErrIntIn},
		},
		{
			name:        "non-struct input",
			in:          "not a struct",
			expectedErr: ErrNotStruct,
		},
		{
			name:        "non-struct int input",
			in:          42,
			expectedErr: ErrNotStruct,
		},
		{
			name: "multiple errors accumulate",
			in: User{
				ID:     "short",
				Name:   "X",
				Age:    5,
				Email:  "bad",
				Role:   "unknown",
				Phones: []string{"1"},
			},
			wantErrs: []error{ErrStringLen, ErrIntMin, ErrStringRegexp, ErrStringIn, ErrStringLen},
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
	t.Run("unknown rule returns program error", func(t *testing.T) {
		type S struct {
			F string `validate:"unknown:val"`
		}
		err := Validate(S{F: "hello"})
		require.ErrorIs(t, err, ErrInvalidTag)
	})

	t.Run("bad len param returns program error", func(t *testing.T) {
		type S struct {
			F string `validate:"len:abc"`
		}
		err := Validate(S{F: "hello"})
		require.ErrorIs(t, err, ErrInvalidTag)
	})

	t.Run("bad min param returns program error", func(t *testing.T) {
		type S struct {
			N int `validate:"min:abc"`
		}
		err := Validate(S{N: 5})
		require.ErrorIs(t, err, ErrInvalidTag)
	})

	t.Run("bad regexp returns program error", func(t *testing.T) {
		type S struct {
			F string `validate:"regexp:[invalid"`
		}
		err := Validate(S{F: "hello"})
		require.ErrorIs(t, err, ErrInvalidTag)
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
