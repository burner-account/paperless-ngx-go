package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

// technical config
const (
	TEST_LOCALHOST_PORT          int           = 8000
	TEST_REQUEST_TIMEOUT         time.Duration = 2 * time.Second
	TEST_DOCUMENT_UPLOAD_TIMEOUT time.Duration = 10 * time.Second
)

// paperless-ngx config
const (
	TEST_USER                string = "test"
	TEST_PASSWORD            string = "test"
	TEST_EMAIL               string = "test@localhost"
	TEST_TIMEZONE            string = "Europe/Berlin"
	TEST_MAIN_LANGUAGE       string = "deu"
	TEST_SUPPORTED_LANGUAGES string = "eng"
)

func dockerCompose() string {
	return fmt.Sprintf(`
services:
  broker:
    image: docker.io/library/redis:8
    restart: unless-stopped
    volumes:
      - redisdata:/data
  db:
    image: docker.io/library/postgres:18
    restart: unless-stopped
    volumes:
      - pgdata:/var/lib/postgresql
    environment:
      POSTGRES_DB: paperless
      POSTGRES_USER: paperless
      POSTGRES_PASSWORD: paperless
  webserver:
    image: ghcr.io/paperless-ngx/paperless-ngx:latest
    restart: unless-stopped
    depends_on:
      - db
      - broker
      - gotenberg
      - tika
    ports:
      - "%d:8000"
    volumes:
      - data:/usr/src/paperless/data
      - media:/usr/src/paperless/media
      - ./export:/usr/src/paperless/export
      - ./consume:/usr/src/paperless/consume
    environment:
      PAPERLESS_REDIS: redis://broker:6379
      PAPERLESS_DBHOST: db
      PAPERLESS_TIKA_ENABLED: 1
      PAPERLESS_TIKA_GOTENBERG_ENDPOINT: http://gotenberg:3000
      PAPERLESS_TIKA_ENDPOINT: http://tika:9998
  gotenberg:
    image: docker.io/gotenberg/gotenberg:8.24
    restart: unless-stopped
    command:
      - "gotenberg"
      - "--chromium-disable-javascript=true"
      - "--chromium-allow-list=file:///tmp/.*"
  tika:
    image: docker.io/apache/tika:latest
    restart: unless-stopped
volumes:
  data:
  media:
  pgdata:
  redisdata:
`, TEST_LOCALHOST_PORT)
}

func initializeSuperUser(ctx context.Context, stack *compose.DockerCompose) error {
	wsContainer, err := stack.ServiceContainer(ctx, "webserver")
	if err != nil {
		return fmt.Errorf("failed to access container: %w", err)
	}
	_, _, err = wsContainer.Exec(
		ctx,
		[]string{
			"/usr/local/bin/python3",
			"manage.py",
			"createsuperuser",
			"--noinput", "--username", TEST_USER, "--email", TEST_EMAIL,
		},
		exec.WithEnv([]string{fmt.Sprintf("DJANGO_SUPERUSER_PASSWORD=%s", TEST_PASSWORD)}),
	)
	if err != nil {
		return fmt.Errorf("failed to init superuser: %w", err)
	}
	return nil
}

func runRecoverably(ctx context.Context, m *testing.M, stack *compose.DockerCompose) (exitcode int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from: %v", r)
		}
	}()
	exitcode = 1

	err := initializeSuperUser(ctx, stack)
	if err != nil {
		log.Printf("❌ failed to initialize superuser: %v\n", err)
		return
	}

	log.Println("⏳ seeding pdf documents")
	ctx, cancel := context.WithTimeout(ctx, 3*TEST_DOCUMENT_UPLOAD_TIMEOUT)
	err = seedDocuments(ctx)
	cancel()
	if err != nil {
		log.Printf("❌ failed to seed documents: %v\n", err)
		return
	}
	log.Println("✅ finished seeding pdf documents")
	log.Println("✨ running tests...")
	exitcode = m.Run()
	return
}

func TestMain(m *testing.M) {
	stack, err := compose.NewDockerComposeWith(
		compose.WithStackReaders(
			strings.NewReader(dockerCompose()),
		),
	)
	if err != nil {
		log.Printf("❌ failed to create stack: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stack.
		WithEnv(map[string]string{
			"PAPERLESS_URL":           baseURL(),
			"PAPERLESS_TIME_ZONE":     TEST_TIMEZONE,
			"PAPERLESS_OCR_LANGUAGE":  TEST_MAIN_LANGUAGE,
			"PAPERLESS_OCR_LANGUAGES": TEST_SUPPORTED_LANGUAGES,
		}).
		WaitForService(
			"webserver",
			wait.ForHealthCheck(),
		).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Printf("❌ failed to start stack: %v\n", err)
		os.Exit(1)
	}

	exitCode := runRecoverably(ctx, m, stack)

	err = stack.Down(
		ctx,
		compose.RemoveOrphans(true),
		compose.RemoveVolumes(true),
		compose.RemoveImagesLocal,
	)
	if err != nil {
		log.Printf("❌ failed to stop stack: %v\n", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
}
