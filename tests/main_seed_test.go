package tests

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/burner-account/paperless-ngx-go"
	"golang.org/x/sync/errgroup"
)

type seedFile struct {
	Filename string
	Title    string
}

var testdataDocuments []seedFile = []seedFile{
	{
		Filename: "./testdata/seed/asimov-wikipedia.pdf",
		Title:    "wikipedia entry on Isaac Asimov (partial)",
	},
	{
		Filename: "./testdata/seed/rocket-wikipedia.pdf",
		Title:    "wikipedia entry on Stephenson's Rocket (partial)",
	},
}

func seedDocuments(ctx context.Context) error {
	client, err := makeClient()
	if err != nil {
		return fmt.Errorf("failed to create client (seed documents): %w", err)
	}

	eg := errgroup.Group{}

	for _, td := range testdataDocuments {
		eg.Go(func() error {
			taskID, err := uploadDocument(ctx, client, td.Filename, td.Title)
			if err != nil {
				return err
			}
			return waitForTask(ctx, client, taskID)
		})
	}
	return eg.Wait()
}

func uploadDocument(ctx context.Context, client paperless.XClient, uploadFilepath string, title string) (string, error) {
	uploadResp, err := client.DocumentsPostDocumentCreateWithBodyWithResponse(
		ctx,
		uploadFilepath,
		&paperless.DocumentCreate{
			Title: paperless.P(title),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}
	if http.StatusOK != uploadResp.HTTPResponse.StatusCode {
		return "", fmt.Errorf("invalid response code (upload document): %d", uploadResp.HTTPResponse.StatusCode)
	}
	if uploadResp.JSON200 == nil {
		return "", fmt.Errorf("response task id nil (upload document)")
	}
	return *uploadResp.JSON200, nil
}

func waitForTask(ctx context.Context, client paperless.XClient, taskID string) error {
	ticker := time.NewTicker(TEST_REQUEST_TIMEOUT + 500*time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for task id '%s' timed out", taskID)
		case <-ticker.C:
			innerCtx, cancel := context.WithTimeout(context.Background(), TEST_REQUEST_TIMEOUT)
			task, err := fetchTask(innerCtx, client, taskID)
			cancel()
			if err != nil {
				return fmt.Errorf("waiting for task id '%s' failed: %w", taskID, err)
			}
			if containsTaskStatus(taskErrorStatus, *task.Status) {
				return fmt.Errorf("task with id '%s' has status: %s", taskID, *task.Status)
			}
			if paperless.StatusEnumSUCCESS == *task.Status {
				return nil
			}
		}
	}
}

func fetchTask(ctx context.Context, client paperless.XClient, taskID string) (*paperless.TasksView, error) {
	listParams := &paperless.TasksListParams{
		TaskId: &taskID,
	}
	taskResp, err := client.TasksListWithResponse(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
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
