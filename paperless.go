package paperless

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

const defaultAPIVersion int = 9

//go:generate go tool oapi-codegen -config cfg.yaml api.yaml

func P[T any](v T) *T {
	return &v
}

func MakeTokenAuthRequestEditor(token string) RequestEditorFn {
	authHeaderValue := fmt.Sprintf("Token %s", token)
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", authHeaderValue)
		return nil
	}
}

func MakeBasicAuthRequestEditor(user, password string) RequestEditorFn {
	credentials := fmt.Sprintf("%s:%s", user, password)
	authHeaderValue := fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString([]byte(credentials)),
	)
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", authHeaderValue)
		return nil
	}
}

func MakeAPIVersionRequestEditor(version int) RequestEditorFn {
	acceptHeaderValue := fmt.Sprintf("application/json; version=%d", version)
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Accept", acceptHeaderValue)
		return nil
	}
}
