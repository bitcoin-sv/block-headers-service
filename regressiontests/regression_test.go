package regressiontests

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestApplicationIntegration(t *testing.T) {
	ctx := context.Background()

	postgresContainer, mappedPort := startPostgresContainer(ctx, t)
	defer postgresContainer.Terminate(ctx)
	fmt.Printf("Connect to PostgreSQL on localhost:%s\n", mappedPort)

	cmd := exec.Command("go", "run", "../cmd/main.go", "-C", "../regressiontests/postgres.config.yaml")
	cmd.Env = append(os.Environ(), "BHS_DB_POSTGRES_PORT="+mappedPort)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start the application: %v", err)
	}

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Failed to kill application process: %v", err)
		}
	}()

	appExitSignal := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		appExitSignal <- err
	}()

	select {
	case err := <-appExitSignal:
		if err != nil {
			t.Fatalf("Application exited unexpectedly: %v", err)
		}
	case <-time.After(1 * time.Minute):
	}

	resp, err := http.Get("http://localhost:8081/status")
	if err != nil {
		t.Fatalf("Failed to make HTTP request to the application: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func startPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "bhs",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Could not start postgres container: %v", err)
	}

	time.Sleep(10 * time.Second) // Wait for the container to be ready

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}
	return postgresContainer, mappedPort.Port()
}
