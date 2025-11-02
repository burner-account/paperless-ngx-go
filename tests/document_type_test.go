package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/burner-account/paperless-ngx-go"
	"github.com/stretchr/testify/require"
)

func TestDocumentTypeCRUD(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)
	var (
		docTypeName string = randStr(32)
		// stores document type returned by API calls
		docType *paperless.DocumentType
	)
	// one timeout for all API requests
	ctx, cancel := context.WithTimeout(context.Background(), 6*TEST_REQUEST_TIMEOUT)
	defer cancel()

	// create document type to toy with
	createParam := paperless.DocumentTypesCreateJSONRequestBody{
		Name:              docTypeName,
		MatchingAlgorithm: paperless.P(paperless.MatchingAlgorithmN6),
		IsInsensitive:     paperless.P(true),
	}
	createResp, err := client.DocumentTypesCreateWithResponse(ctx, createParam)

	require.NoError(err, "failed to create document type")
	require.Equal(
		http.StatusCreated,
		createResp.HTTPResponse.StatusCode,
		"invalid response code (create document type)",
	)
	require.NotNil(createResp.JSON201, "response json nil (create document type)")
	require.NotNil(createResp.JSON201.Id, "id nil (create document type)")
	docType = createResp.JSON201

	// list document types
	listResp, err := client.DocumentTypesListWithResponse(ctx, nil)

	require.NoError(err, "failed to list document types")
	require.Equal(
		http.StatusOK,
		listResp.HTTPResponse.StatusCode,
		"invalid response code (list document types)",
	)
	require.NotNil(listResp.JSON200, "response json nil (get document type)")
	require.Greater(listResp.JSON200.Count, 0, "document type count")

	// use GET functionality of document type
	getResp, err := client.DocumentTypesRetrieveWithResponse(ctx, *docType.Id, nil)

	require.NoError(err, "failed to get document type")
	require.Equal(
		http.StatusOK,
		getResp.HTTPResponse.StatusCode,
		"invalid response code (get document type)",
	)
	require.NotNil(getResp.JSON200, "response json nil (get document type)")
	require.NotNil(getResp.JSON200.Id, "id nil (get document type)")
	require.Equal(
		*docType.Id,
		*getResp.JSON200.Id,
		"document type id",
	)
	docType = getResp.JSON200

	// use PATCH functionality of document type
	patchParam := paperless.DocumentTypesPartialUpdateJSONRequestBody{
		Name: paperless.P(fmt.Sprintf("new-%s", docType.Name)),
	}
	patchResp, err := client.DocumentTypesPartialUpdateWithResponse(ctx, *docType.Id, patchParam)

	require.NoError(err, "failed to patch document type")
	require.Equal(
		http.StatusOK,
		patchResp.HTTPResponse.StatusCode,
		"invalid response code (patch document type)",
	)
	require.NotNil(patchResp.JSON200, "response json nil (patch document type)")
	require.NotNil(patchResp.JSON200.Id, "id nil (patch document type)")
	require.Equal(
		*docType.Id,
		*patchResp.JSON200.Id,
		"document type id",
	)
	docType = patchResp.JSON200

	// use PUT functionality of document type
	putParam := paperless.DocumentTypesUpdateJSONRequestBody{
		IsInsensitive:     docType.IsInsensitive,
		Match:             docType.Match,
		MatchingAlgorithm: docType.MatchingAlgorithm,
		Name:              docType.Name,
		Owner:             docType.Owner,
		SetPermissions:    docType.Permissions,
	}
	putResp, err := client.DocumentTypesUpdateWithResponse(ctx, *docType.Id, putParam)

	require.NoError(err, "failed to put document type")
	require.Equal(
		http.StatusOK,
		putResp.HTTPResponse.StatusCode,
		"invalid response code (put document type)",
	)
	require.NotNil(putResp.JSON200, "response json nil (put document type)")
	require.NotNil(putResp.JSON200.Id, "id nil (put document type)")
	require.Equal(
		*docType.Id,
		*putResp.JSON200.Id,
		"document type id",
	)
	docType = putResp.JSON200

	// delete the document type previously created
	destroyResp, err := client.DocumentTypesDestroyWithResponse(ctx, *docType.Id)

	require.NoError(err, "failed to delete document type")
	require.Equal(
		http.StatusNoContent,
		destroyResp.HTTPResponse.StatusCode,
		"invalid response code (delete document type)",
	)
}
