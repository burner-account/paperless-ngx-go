package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatistics(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
	defer cancel()
	statistics, err := client.StatisticsRetrieveWithResponse(ctx)

	require.NoError(err, "failed to get server statistics")
	require.Equal(
		http.StatusOK,
		statistics.HTTPResponse.StatusCode,
		"invalid response code (get server statistics)",
	)
}
