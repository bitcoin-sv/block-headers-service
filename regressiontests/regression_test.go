package regressiontests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"os"

	"io"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/tips"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const whatsonchainAPIURL = "https://api.whatsonchain.com/v1/bsv/main/chain/tips"

type WhatsOnChainForkTip struct {
	Height    int    `json:"height"`
	Hash      string `json:"hash"`
	BranchLen int    `json:"branchlen"`
	Status    string `json:"status"`
}

var currentSyncTime = 1 * time.Minute

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
	case <-time.After(currentSyncTime):
	}

	resp, err := http.Get("http://localhost:8081/status")
	if err != nil {
		t.Fatalf("Failed to make HTTP request to the application: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	wocTipHeight, err := fetchExternalForkHeight(whatsonchainAPIURL)
	if err != nil {
		t.Fatalf("Failed to fetch external fork height: %v", err)
	}

	localTipHeight, err := fetchLocalTip()
	if err != nil {
		t.Fatalf("Failed to fetch local tip height: %v", err)
	}

	if localTipHeight < wocTipHeight {
		t.Errorf("Couldn't sync to proper tip of chain: %d < %d", localTipHeight, wocTipHeight)
	} else {
		t.Logf("Synced to tip of chain: %d", localTipHeight)
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

func fetchExternalForkHeight(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var forks []WhatsOnChainForkTip
	if err := json.NewDecoder(resp.Body).Decode(&forks); err != nil {
		return 0, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	for _, fork := range forks {
		if fork.Status == "active" {
			return fork.Height, nil
		}
	}

	return 0, fmt.Errorf("no active fork found")
}

func fetchLocalTip() (int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8081/api/v1/chain/tip/longest", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.DefaultAppToken)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make HTTP request to application: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.Reader(resp.Body))
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var chainTip tips.TipStateResponse
	if err := json.NewDecoder(resp.Body).Decode(&chainTip); err != nil {
		return 0, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return int(chainTip.Height), nil
}
