package domains

// MerkleRootConfirmationRequestItem is a request type for verification
// of Merkle Roots inclusion in the longest chain.
type MerkleRootConfirmationRequestItem struct {
	MerkleRoot  string `json:"merkleRoot"`
	BlockHeight int32  `json:"blockHeight"`
}

// MerkleRootConfirmationState represents the state of each Merkle Root verification
// process and can be one of three values: Confirmed, Invalid and UnableToVerify.
type MerkleRootConfirmationState string

const (
	// Confirmed state occurs when Merkle Root is found in the longest chain.
	Confirmed MerkleRootConfirmationState = "CONFIRMED"
	// UnableToVerify state occurs when Block Headers Service is behind in synchronization with the longest chain.
	UnableToVerify MerkleRootConfirmationState = "UNABLE_TO_VERIFY"
	// Invalid state occurs when Merkle Root is not found in the longest chain.
	Invalid MerkleRootConfirmationState = "INVALID"
)

// MerkleRootConfirmation is used to confirm the inclusion of
// Merkle Roots in the longest chain.
type MerkleRootConfirmation struct {
	MerkleRoot   string                      `json:"merkleRoot"`
	BlockHeight  int32                       `json:"blockHeight"`
	Hash         string                      `json:"hash,omitempty"`
	Confirmation MerkleRootConfirmationState `json:"confirmation"`
}

// MerkleRootsResponse is the response object that should be returned from the database when
// requested from merkleroots api endpoint
type MerkleRootsResponse struct {
	MerkleRoot  string `json:"merkleRoot"`
	BlockHeight int32  `json:"blockHeight"`
}

// MerkleRootsESKPagedResponse is a paged response model for merkleroots that uses exclusive start key pagination
type MerkleRootsESKPagedResponse = ExclusiveStartKeyPagedResponse[*MerkleRootsResponse, int]
