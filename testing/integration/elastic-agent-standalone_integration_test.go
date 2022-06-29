package integration

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFetch(t *testing.T) {
	err := BuildBinary()
	require.NoError(t, err)
	defer Clean()
}
