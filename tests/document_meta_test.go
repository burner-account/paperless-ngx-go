package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDocumentMeta(t *testing.T) {
	require := require.New(t)
	client := makeTestClient(t)

	uploadFilepath := "./testdata/test-01.pdf"

	ctx, cancel := context.WithTimeout(context.Background(), 5*TEST_DOCUMENT_UPLOAD_TIMEOUT)
	err := client.WaitForDocumentUpload(
		ctx,
		uploadFilepath,
		"Testdokument 01",
		time.Now(),
		[]int{},
	)
	cancel()
	require.NoError(err, "failed to upload document")

	ctx, cancel = context.WithTimeout(context.Background(), TEST_DOCUMENT_UPLOAD_TIMEOUT)
	docs, err := client.GetAllDocuments(ctx)
	cancel()
	require.NoError(err, "failed to fetch document list")
	require.Greater(len(docs), 0, "no document found")

	ctx, cancel = context.WithTimeout(context.Background(), TEST_DOCUMENT_UPLOAD_TIMEOUT)
	metaResp, err := client.DocumentsMetadataRetrieveWithResponse(ctx, *docs[0].Id)
	cancel()
	require.NoError(err, "failed to fetch document metadata")
	require.NotNil(metaResp.JSON200)
}
