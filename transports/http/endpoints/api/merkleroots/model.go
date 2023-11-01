package merkleroots

import (
	"github.com/bitcoin-sv/pulse/domains"
)

// MerkleRootConfirmation is a confirmation
// of merkle roots inclusion in the longest chain.
type MerkleRootConfirmation struct {
	Hash         string                              `json:"blockHash"`
	BlockHeight  int32                               `json:"blockHeight"`
	MerkleRoot   string                              `json:"merkleRoot"`
	Confirmation domains.MerkleRootConfirmationState `json:"confirmation"`
}

// MerkleRootsConfirmationsResponse is an API response for confirming
// merkle roots inclusion in the longest chain.
type MerkleRootsConfirmationsResponse struct {
	ConfirmationState domains.MerkleRootConfirmationState `json:"confirmationState"`
	Confirmations     []MerkleRootConfirmation            `json:"confirmations"`
}

// newMerkleRootConfirmationcreates a new merkleRootConfirmation
// object from domain's MerkleRootConfirmation object.
func newMerkleRootConfirmation(
	merkleConfm *domains.MerkleRootConfirmation,
) MerkleRootConfirmation {
	return MerkleRootConfirmation{
		Hash:         merkleConfm.Hash,
		BlockHeight:  merkleConfm.BlockHeight,
		MerkleRoot:   merkleConfm.MerkleRoot,
		Confirmation: merkleConfm.Confirmation,
	}
}

// mapToMerkleRootsConfirmationsResponses converts a slice of domain's
// MerkleRootConfirmation objects to merkleRootConfirmationRespose.
func mapToMerkleRootsConfirmationsResponses(
	merkleConfms []*domains.MerkleRootConfirmation,
) MerkleRootsConfirmationsResponse {
	mrcfs := make([]MerkleRootConfirmation, 0)

	confirmationState := domains.Confirmed

	for _, merkleConfm := range merkleConfms {
		mrcfs = append(mrcfs, newMerkleRootConfirmation(merkleConfm))
		if convertState(confirmationState) < convertState(merkleConfm.Confirmation) {
			confirmationState = merkleConfm.Confirmation
		}
	}

	return MerkleRootsConfirmationsResponse{
		ConfirmationState: confirmationState,
		Confirmations:     mrcfs,
	}
}

func convertState(s domains.MerkleRootConfirmationState) int {
	switch s {
	case domains.Confirmed:
		return 0
	case domains.UnableToVerify:
		return 1
	case domains.Invalid:
		return 2
	default:
		return 2
	}
}
