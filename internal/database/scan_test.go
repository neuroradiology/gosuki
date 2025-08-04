package database

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestScanXxhsum(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  xxhashsum
		err   string
	}{
		{
			name:  "nil",
			input: nil,
			want:  0,
			err:   "",
		},
		{
			name:  "uint64",
			input: uint64(123),
			want:  123,
			err:   "",
		},
		{
			name:  "string valid",
			input: "456",
			want:  456,
			err:   "",
		},
		{
			name:  "string invalid",
			input: "abc",
			want:  0,
			err:   "cannot parse uint64 from string \"abc\"",
		},
		{
			name:  "byte slice valid",
			input: []byte("789"),
			want:  789,
			err:   "",
		},
		{
			name:  "byte slice invalid",
			input: []byte("def"),
			want:  0,
			err:   "cannot parse uint64 from []byte [100 101 102]",
		},
		{
			name:  "other type",
			input: true,
			want:  0,
			err:   "cannot convert to uint64 true",
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
			err:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h xxhashsum
			err := h.Scan(tt.input)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tt.err)
			}
			require.Equal(t, tt.want, h)
		})
	}
}

func TestNodeUUID_Scan_Nil(t *testing.T) {
	var id UUID
	err := id.Scan(nil)
	require.NoError(t, err)
	require.Equal(t, UUID(uuid.Nil), id)
}

func TestNodeUUID_Scan_ValidBytes(t *testing.T) {
	u, err := uuid.NewV4()
	require.NoError(t, err)
	bytes := u.Bytes()
	var id UUID
	err = id.Scan(bytes)
	require.NoError(t, err)
	require.Equal(t, UUID(u), id)
}

func TestNodeUUID_Scan_InvalidType(t *testing.T) {
	var id UUID
	err := id.Scan("not a byte slice")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot parse uuid.UUID from")
}

func TestNodeUUID_Scan_InvalidBytes(t *testing.T) {
	bytes := make([]byte, 15)
	var id UUID
	err := id.Scan(bytes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not parse uuid.UUID")
}
