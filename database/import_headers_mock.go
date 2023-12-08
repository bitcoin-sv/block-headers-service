package database

import (
	"github.com/bitcoin-sv/pulse/database/sql"
	"github.com/bitcoin-sv/pulse/domains"
	"github.com/bitcoin-sv/pulse/internal/chaincfg/chainhash"
)

type MockHeaders struct {
	db *sql.HeadersDb
}

// AddHeaderToDatabase implements the corresponding method in MockHeaders.
func (m *MockHeaders) AddHeaderToDatabase(header domains.BlockHeader) error {
	// Implement mock behavior, e.g., just return nil for testing
	return nil
}

// UpdateState implements the corresponding method in MockHeaders.
func (m *MockHeaders) UpdateState(hashes []chainhash.Hash, state domains.HeaderState) error {
	// Implement mock behavior, e.g., just return nil for testing
	return nil
}

// GetHeaderByHeight implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetHeaderByHeight(height int32) (*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a predefined header for testing
	return &domains.BlockHeader{}, nil
}

// GetHeaderByHeightRange implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetHeaderByHeightRange(from int, to int) ([]*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a slice of predefined headers for testing
	headers := []*domains.BlockHeader{
		&domains.BlockHeader{},
		&domains.BlockHeader{},
		// Add more headers as needed
	}
	return headers, nil
}

// GetLongestChainHeadersFromHeight implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetLongestChainHeadersFromHeight(height int32) ([]*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a slice of predefined headers for testing
	headers := []*domains.BlockHeader{
		&domains.BlockHeader{},
		&domains.BlockHeader{},
		// Add more headers as needed
	}
	return headers, nil
}

// GetStaleChainHeadersBackFrom implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetStaleChainHeadersBackFrom(hash string) ([]*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a slice of predefined headers for testing
	headers := []*domains.BlockHeader{
		&domains.BlockHeader{},
		&domains.BlockHeader{},
		// Add more headers as needed
	}
	return headers, nil
}

// GetCurrentHeight implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetCurrentHeight() (int, error) {
	// Implement mock behavior, e.g., return a predefined height for testing
	return 42, nil
}

// GetHeadersCount implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetHeadersCount() (int, error) {
	// Implement mock behavior, e.g., return a predefined count for testing
	return 100, nil
}

// GetHeaderByHash implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetHeaderByHash(hash string) (*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a predefined header for testing
	return &domains.BlockHeader{}, nil
}

// GetMerkleRootsConfirmations implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetMerkleRootsConfirmations(request []domains.MerkleRootConfirmationRequestItem, maxBlockHeightExcess int) ([]*domains.MerkleRootConfirmation, error) {
	// Implement mock behavior, e.g., return a slice of predefined MerkleRootConfirmations for testing
	confirmations := []*domains.MerkleRootConfirmation{
		{},
		{},
		// Add more confirmations as needed
	}
	return confirmations, nil
}

// GenesisExists implements the corresponding method in MockHeaders.
func (m *MockHeaders) GenesisExists() bool {
	// Implement mock behavior, e.g., return a predefined value for testing
	return true
}

// GetPreviousHeader implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetPreviousHeader(hash string) (*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a predefined header for testing
	return &domains.BlockHeader{}, nil
}

// GetTip implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetTip() (*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a predefined tip for testing
	return &domains.BlockHeader{}, nil
}

// GetAllTips implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetAllTips() ([]*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a slice of predefined tips for testing
	headers := []*domains.BlockHeader{
		&domains.BlockHeader{},
		&domains.BlockHeader{},
		// Add more headers as needed
	}
	return headers, nil
}

// GetAncestorOnHeight implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetAncestorOnHeight(hash string, height int32) (*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a predefined header for testing
	return &domains.BlockHeader{}, nil
}

// GetChainBetweenTwoHashes implements the corresponding method in MockHeaders.
func (m *MockHeaders) GetChainBetweenTwoHashes(low string, high string) ([]*domains.BlockHeader, error) {
	// Implement mock behavior, e.g., return a slice of predefined headers for testing
	headers := []*domains.BlockHeader{
		&domains.BlockHeader{},
		&domains.BlockHeader{},
		// Add more headers as needed
	}
	return headers, nil
}

func (m *MockHeaders) AddMultipleHeadersToDatabase(p0 []domains.BlockHeader) error {
	return nil
}

// AddHeaderToDatabasePointerReceiver implements the corresponding method in MockHeaders.
func (m *MockHeaders) AddHeaderToDatabasePointerReceiver(header *domains.BlockHeader) error {
	// Implement mock behavior, e.g., just return nil for testing
	return nil
}
