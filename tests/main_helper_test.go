package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/burner-account/paperless-ngx-go"
)

func baseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", TEST_LOCALHOST_PORT)
}

func makeClient() (paperless.XClient, error) {
	return paperless.NewXClientWithCredentials(
		baseURL(),
		TEST_USER,
		TEST_PASSWORD,
	)
}

func makeTestClient(t *testing.T) paperless.XClient {
	client, err := makeClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return client
}

var pool = []rune("abcdef1234567890")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = pool[rand.Intn(len(pool))]
	}
	return string(b)
}

var (
	taskErrorStatus []paperless.StatusEnum = []paperless.StatusEnum{
		paperless.StatusEnumFAILURE,
		paperless.StatusEnumREVOKED,
	}
)

func containsTaskStatus(arr []paperless.StatusEnum, status paperless.StatusEnum) bool {
	for _, test := range arr {
		if test == status {
			return true
		}
	}
	return false
}
