package paperless

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
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
		output["created"] = d.Created.Format(APIDateTimeFormat)
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

func (x XClient) DocumentsPostDocumentCreateWithBodyWithResponse(
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

func (x XClient) UploadDocument(ctx context.Context, filepath, title string, created time.Time, tagIDs []int) (string, error) {
	var tags []string
	if len(tagIDs) > 0 {
		tags = make([]string, len(tagIDs))
		for idx, tagID := range tagIDs {
			tags[idx] = strconv.Itoa(tagID)
		}
	}
	docResp, err := x.DocumentsPostDocumentCreateWithBodyWithResponse(
		ctx,
		filepath,
		&DocumentCreate{
			Title:   P(title),
			Created: P(created),
			Tags:    tags,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}
	if docResp.JSON200 == nil {
		return "", fmt.Errorf("missing json response, not task waiting for '%s'", filepath)

	}
	return *docResp.JSON200, nil
}

func (x XClient) WaitForDocumentUpload(ctx context.Context, filepath, title string, created time.Time, tagIDs []int) error {
	taskId, err := x.UploadDocument(ctx, filepath, title, created, tagIDs)
	if err != nil {
		return err
	}
	err = x.WaitForTask(ctx, taskId, 2*time.Second)
	if err != nil {
		err = fmt.Errorf("failed to upload '%s': %w", filepath, err)
	}
	return err

}

func (x XClient) FetchTask(ctx context.Context, taskID string) (*TasksView, error) {
	taskResp, err := x.TasksListWithResponse(ctx, &TasksListParams{
		TaskId: &taskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve task: %w", err)
	}
	if taskResp.JSON200 == nil {
		return nil, fmt.Errorf("response json nil (list tasks)")
	}
	for _, task := range taskResp.JSON200 {
		if task.TaskId == taskID {
			return &task, nil
		}
	}
	return nil, fmt.Errorf("task not found")
}

func (x XClient) WaitForTask(ctx context.Context, taskID string, pollInterval time.Duration) error {
	if pollInterval < 2000*time.Millisecond {
		pollInterval = 2000 * time.Millisecond
	}
	ticker := time.NewTicker(pollInterval)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for task id '%s' timed out", taskID)
		case <-ticker.C:
			innerCtx, cancel := context.WithTimeout(context.Background(), pollInterval-500*time.Millisecond)
			task, err := x.FetchTask(innerCtx, taskID)
			cancel()
			if err != nil {
				continue
				// return fmt.Errorf("waiting for task id '%s' failed: %w", taskID, err)
			}
			if *task.Status == StatusEnumFAILURE || *task.Status == StatusEnumREVOKED {
				return fmt.Errorf("task with id '%s' has status: %s", taskID, *task.Status)
			}
			if StatusEnumSUCCESS == *task.Status {
				return nil
			}
		}
	}
}

func (x XClient) GetAllDocuments(ctx context.Context) ([]Document, error) {
	docResp, err := x.DocumentsListWithResponse(ctx, &DocumentsListParams{
		PageSize: P(9999999),
	})
	if err != nil {
		return nil, fmt.Errorf("GetAllDocumentss failed: %w", err)
	}
	if docResp.JSON200 == nil {
		return nil, fmt.Errorf("GetAllDocumentss failed: document json response nil")
	}
	return docResp.JSON200.Results, nil
}
