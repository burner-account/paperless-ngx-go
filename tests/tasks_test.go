package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/burner-account/paperless-ngx-go"
	"github.com/stretchr/testify/require"
)

func TestTasksList(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
	defer cancel()

	var tasksResp *paperless.TasksListHTTPResponse
	tasksResp, err := client.TasksListWithResponse(ctx, nil)

	require.NoError(err, "failed to get tasks list")
	require.Equal(
		http.StatusOK,
		tasksResp.HTTPResponse.StatusCode,
		"invalid response code (get tasks list)",
	)
	require.NotNil(tasksResp.JSON200, "response json nil (get tasks list)")

}
