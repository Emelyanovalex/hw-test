package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func TestCopy(t *testing.T) {
	cases := []struct {
		offset   int64
		limit    int64
		expected string
	}{
		{0, 0, "out_offset0_limit0.txt"},
		{0, 10, "out_offset0_limit10.txt"},
		{0, 1000, "out_offset0_limit1000.txt"},
		{0, 10000, "out_offset0_limit10000.txt"},
		{100, 1000, "out_offset100_limit1000.txt"},
		{6000, 1000, "out_offset6000_limit1000.txt"},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			dst, err := os.CreateTemp("", "copy_test_*")
			require.NoError(t, err)
			dst.Close()
			defer os.Remove(dst.Name())

			err = Copy(filepath.Join("testdata", "input.txt"), dst.Name(), tc.offset, tc.limit)
			require.NoError(t, err)

			got := readFile(t, dst.Name())
			want := readFile(t, filepath.Join("testdata", tc.expected))
			require.Equal(t, want, got)
		})
	}
}

func TestCopyOffsetExceedsFileSize(t *testing.T) {
	dst, err := os.CreateTemp("", "copy_test_*")
	require.NoError(t, err)
	dst.Close()
	defer os.Remove(dst.Name())

	err = Copy(filepath.Join("testdata", "input.txt"), dst.Name(), 1<<32, 0)
	require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
}

func TestCopyUnsupportedFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no device files on Windows")
	}

	dst, err := os.CreateTemp("", "copy_test_*")
	require.NoError(t, err)
	dst.Close()
	defer os.Remove(dst.Name())

	err = Copy("/dev/urandom", dst.Name(), 0, 0)
	require.ErrorIs(t, err, ErrUnsupportedFile)
}
