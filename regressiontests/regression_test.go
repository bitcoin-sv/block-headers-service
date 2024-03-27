package regressiontests

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/headers"
	"github.com/bitcoin-sv/block-headers-service/transports/http/endpoints/api/merkleroots"
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

var dbEngine string

func init() {
	flag.StringVar(&dbEngine, "dbEngine", "postgres", "The database engine to use in tests (postgres or sqlite)")
}

func TestApplicationIntegration(t *testing.T) {
	flag.Parse()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	timeoutTimer := time.NewTimer(2 * time.Minute)
	defer timeoutTimer.Stop()

	ctx := context.Background()

	buildCmd := exec.Command("go", "build", "-o", "../cmd/bhsExecutable", "../cmd/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build the application for tests: %v", err)
	}
	defer func() {
		err := os.Remove("../cmd/bhsExecutable")
		if err != nil {
			t.Logf("Couldn't remove executable files %v", err)
		}
	}()

	var cmd *exec.Cmd

	switch dbEngine {
	case "postgres":
		postgresContainer, mappedPort := startPostgresContainer(ctx, t)
		defer func(postgresContainer testcontainers.Container, ctx context.Context) {
			err := postgresContainer.Terminate(ctx)
			if err != nil {
				t.Logf("Couldn't terminate postgres container %v", err)
			}
		}(postgresContainer, ctx)
		cmd = exec.Command("../cmd/bhsExecutable", "-C", "../regressiontests/postgres.config.yaml")
		cmd.Env = append(os.Environ(), "BHS_DB_POSTGRES_PORT="+mappedPort)

	case "sqlite":
		cmd = exec.Command("../cmd/bhsExecutable", "-C", "../regressiontests/sqlite.config.yaml")

		defer func() {
			err := os.Remove("../data/blockheaders.db")
			if err != nil {
				t.Logf("Warning: Failed to remove SQLite database file: %v", err)
			} else {
				t.Log("SQLite database file removed successfully.")
			}
		}()

	default:
		t.Fatalf("Unsupported database engine: %s", dbEngine)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	defer func() {
		t.Log("Killing application process...")
		t.Logf("pid: %d", cmd.Process.Pid)
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to kill application process: %v", err)
		}
	}()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start the application: %v", err)
	}

	appExitSignal := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		appExitSignal <- err
	}()

out:
	for {
		select {
		case err := <-appExitSignal:
			if err != nil {
				t.Fatalf("Application exited unexpectedly: %v", err)
			}
		case <-timeoutTimer.C:
			t.Fatalf("Test timed out after 2 minutes without passing all checks.")
		case <-ticker.C:
			resp, err := http.Get("http://localhost:8080/status")
			if err != nil {
				t.Logf("Failed to make HTTP request to the application: %v - next attempt in 10s", err)
				continue
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Logf("Couldn't close http status request body %v", err)
				}
			}(resp.Body)

			if resp.StatusCode != 200 {
				t.Logf("Expected status code 200, got %d - next attempt in 10s", resp.StatusCode)
				continue
			}

			t.Log("HTTP server is up and running")

			wocTipHeight, err := fetchExternalForkHeight(whatsonchainAPIURL)
			if err != nil {
				t.Logf("Failed to fetch external fork height: %v", err)
				continue
			}

			localTipHeight, err := fetchLocalTip()
			if err != nil {
				t.Logf("Failed to fetch local tip height: %v", err)
				continue
			}

			if localTipHeight < wocTipHeight {
				t.Logf("Couldn't sync to proper tip of chain - next attempt in 10s: %d < %d", localTipHeight, wocTipHeight)
				continue
			} else {
				t.Logf("Synced to tip of chain: %d", localTipHeight)
				break out
			}
		}
	}

	t.Log("Comparing synced data with known checkpoints...")

	verifyHeaders(t)

	t.Log("Headers checkpoint comparison passed successfully 🎉")

	t.Log("Verifying merkle roots...")

	verifyMerkleRoots(t, fixtures)

	t.Log("Merkle roots verification passed successfully 🎉")
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
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/chain/tip/longest", nil)
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

func fetchBlockHeader(hash string) (*headers.BlockHeaderStateResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/chain/header/state/"+hash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.DefaultAppToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to application: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var blockHeader headers.BlockHeaderStateResponse
	if err := json.NewDecoder(resp.Body).Decode(&blockHeader); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return &blockHeader, nil
}

func verifyHeaders(t *testing.T) {
	for _, checkpoint := range chaincfg.MainNetCheckpoints {
		blockHeader, err := fetchBlockHeader(checkpoint.Hash.String())
		if err != nil {
			t.Fatalf("Failed to fetch block header for hash %s: %v", checkpoint.Hash.String(), err)
		}

		if blockHeader.Height != checkpoint.Height || blockHeader.State != "LONGEST_CHAIN" {
			t.Fatalf("Checkpoint height mismatch for hash %s: expected %d, got %d", checkpoint.Hash, checkpoint.Height, blockHeader.Height)
		}
	}
}

func verifyMerkleRoots(t *testing.T, fixtures []merkleRootFixtures) {
	confirmations := fetchMerkleRootConfirmations(fixtures, t)

	for _, fixture := range fixtures {
		found := false
		for _, confirmation := range confirmations {
			if fixture.MerkleRoot == confirmation.MerkleRoot && fixture.Height == confirmation.BlockHeight && confirmation.Confirmation == domains.Confirmed {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Merkle root %s at height %d not confirmed as expected.", fixture.MerkleRoot, fixture.Height)
		}
	}
}

func fetchMerkleRootConfirmations(fixtures []merkleRootFixtures, t *testing.T) []merkleroots.MerkleRootConfirmation {
	jsonData, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/chain/merkleroot/verify", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.DefaultAppToken)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make HTTP request to application: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Logf("Couldn't close merkle root response body %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	var confirmations merkleroots.MerkleRootsConfirmationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&confirmations); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	return confirmations.Confirmations
}
