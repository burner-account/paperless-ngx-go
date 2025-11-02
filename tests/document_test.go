package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/burner-account/paperless-ngx-go"
	"github.com/stretchr/testify/require"
)

func TestDocumentUpload(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	uploadFilepath := "./testdata/squirrel-wikipedia.pdf"

	ctx, cancel := context.WithTimeout(context.Background(), TEST_DOCUMENT_UPLOAD_TIMEOUT)
	defer cancel()

	var uploadResp *paperless.DocumentsPostDocumentCreateHTTPResponse

	uploadResp, err := client.DocumentsPostDocumentCreateWithBodyWithResponse(
		ctx,
		uploadFilepath,
		&paperless.DocumentCreate{
			Title: paperless.P("wikipedia about squirrels (partial)"),
		},
	)

	require.NoError(err, "failed to upload document")
	require.Equal(
		http.StatusOK,
		uploadResp.HTTPResponse.StatusCode,
		"invalid response code (upload document)",
	)
	require.NotNil(uploadResp.JSON200, "response json nil (upload document)")
}

func TestAutocomplete(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	autocompleteParams := &paperless.SearchAutocompleteListParams{
		Term:  paperless.P("rock"),
		Limit: paperless.P(10),
	}
	ctx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
	defer cancel()

	autocompleteResp, err := client.SearchAutocompleteListWithResponse(ctx, autocompleteParams)

	require.NoError(err, "failed to get autocomplete")
	require.Equal(
		http.StatusOK,
		autocompleteResp.HTTPResponse.StatusCode,
		"invalid response code (get autocomplete)",
	)
	require.NotNil(autocompleteResp.JSON200, "response json nil (get autocomplete)")
}

func TestSearchRetrieve(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
	defer cancel()

	searchParams := &paperless.SearchRetrieveParams{
		Query: "asimov",
	}
	searchResp, err := client.SearchRetrieveWithResponse(ctx, searchParams)

	require.NoError(err, "failed to get search")
	require.Equal(
		http.StatusOK,
		searchResp.HTTPResponse.StatusCode,
		"invalid response code (get search)",
	)
	require.NotNil(searchResp.JSON200, "response json nil (get search)")
}
