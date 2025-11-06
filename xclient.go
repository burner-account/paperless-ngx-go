package paperless

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"
)

type XClient struct {
	ClientWithResponsesInterface
}

func NewXClient(endpoint string, reqEditors ...RequestEditorFn) (XClient, error) {
	var opts []ClientOption
	for _, editorFn := range reqEditors {
		opts = append(opts, WithRequestEditorFn(editorFn))
	}
	client, err := NewClientWithResponses(
		endpoint,
		opts...,
	)
	if err != nil {
		return XClient{}, fmt.Errorf("could not create client: %w", err)
	}
	return XClient{
		client,
	}, nil
}

func NewXClientWithToken(endpoint, token string) (XClient, error) {
	return NewXClient(
		endpoint,
		MakeAPIVersionRequestEditor(defaultAPIVersion),
		MakeTokenAuthRequestEditor(token),
	)
}

func NewXClientWithCredentials(endpoint, user, password string) (XClient, error) {
	return NewXClient(
		endpoint,
		MakeAPIVersionRequestEditor(defaultAPIVersion),
		MakeBasicAuthRequestEditor(user, password),
	)
}

type DocumentCreate struct {
	Title               *string    `json:"title,omitempty"`
	Created             *time.Time `json:"created,omitempty"`
	DocumentType        *int       `json:"document_type,omitempty"`
	StoragePath         *int       `json:"storage_path,omitempty"`
	Tags                []string   `json:"tags,omitempty"`
	ArchiveSerialNumber *int       `json:"archive_serial_number,omitempty"`
}

func (d *DocumentCreate) Params() map[string]interface{} {
	output := make(map[string]interface{})
	if d != nil && d.Title != nil {
		output["title"] = *d.Title
	}
	if d != nil && d.Created != nil {
		output["created"] = d.Created.Format(apiDateTimeFormat)
	}
	if d != nil && d.DocumentType != nil {
		output["document_type"] = fmt.Sprintf("%d", *d.DocumentType)
	}
	if d != nil && d.StoragePath != nil {
		output["storage_path"] = fmt.Sprintf("%d", *d.StoragePath)
	}
	if d != nil && d.Tags != nil {
		output["tags"] = d.Tags
	}
	if d != nil && d.ArchiveSerialNumber != nil {
		output["archive_serial_number"] = fmt.Sprintf("%d", *d.ArchiveSerialNumber)
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

func (x *XClient) DocumentsPostDocumentCreateWithBodyWithResponse(
	ctx context.Context,
	fullFilepath string,
	optionalData *DocumentCreate,
	reqEditors ...RequestEditorFn,
) (*DocumentsPostDocumentCreateHTTPResponse, error) {
	body, contentType, err := newFileMultipartBody("document", fullFilepath, optionalData.Params())
	if err != nil {
		return nil, err
	}
	return x.ClientWithResponsesInterface.DocumentsPostDocumentCreateWithBodyWithResponse(
		ctx,
		contentType,
		body,
		reqEditors...,
	)
}

func newFileMultipartBody(paramName, path string, params map[string]interface{}) (io.Reader, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return nil, "", err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, "", err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, "", err
	}
	part.Write(fileContents)

	for key, val := range params {
		if str, ok := val.(string); ok {
			_ = writer.WriteField(key, str)
		}
		if sl, ok := val.([]string); ok {
			for _, str := range sl {
				_ = writer.WriteField(key, str)
			}
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
