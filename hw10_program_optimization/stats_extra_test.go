// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat_EmptyInput(t *testing.T) {
	result, err := GetDomainStat(bytes.NewBufferString(""), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

func TestGetDomainStat_CaseInsensitiveEmail(t *testing.T) {
	data := `{"Email":"USER@Example.COM"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"example.com": 1}, result)
}

func TestGetDomainStat_MultipleUsersAggregate(t *testing.T) {
	data := `{"Email":"a@example.com"}
{"Email":"b@example.com"}
{"Email":"c@other.com"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "com")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"example.com": 2, "other.com": 1}, result)
}

func TestGetDomainStat_NoPartialDomainMatch(t *testing.T) {
	data := `{"Email":"user@example.network"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "net")
	require.NoError(t, err)
	require.Equal(t, DomainStat{}, result)
}

func TestGetDomainStat_EmptyEmail(t *testing.T) {
	data := `{"Email":""}
{"Email":"valid@test.org"}`
	result, err := GetDomainStat(bytes.NewBufferString(data), "org")
	require.NoError(t, err)
	require.Equal(t, DomainStat{"test.org": 1}, result)
}
