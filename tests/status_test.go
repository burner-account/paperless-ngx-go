package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
	defer cancel()
	status, err := client.StatusRetrieveWithResponse(ctx)

	require.NoError(err, "failed to get server status")
	require.Equal(
		http.StatusOK,
		status.HTTPResponse.StatusCode,
		"invalid response code (get server status)",
	)
}
