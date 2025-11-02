package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentNotesList(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
	defer cancel()

	notesResp, err := client.DocumentsNotesListWithResponse(ctx, 1, nil)

	require.NoError(err, "failed to get document notes list")
	require.Equal(
		http.StatusOK,
		notesResp.HTTPResponse.StatusCode,
		"invalid response code (get document notes list)",
	)
	require.NotNil(notesResp.JSON200, "response json nil (get document notes list)")

}
