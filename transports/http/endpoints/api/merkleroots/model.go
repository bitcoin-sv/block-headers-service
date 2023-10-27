package merkleroots

import (
	"github.com/bitcoin-sv/pulse/domains"
)

// MerkleRootConfirmation is a confirmation
// of merkle roots inclusion in the longest chain.
type MerkleRootConfirmation struct {
	Hash       string `json:"blockhash"`
	MerkleRoot string `json:"merkleRoot"`
	Confirmed  bool   `json:"confirmed"`
}

// MerkleRootsConfirmationsResponse is an API response for confirming
// merkle roots inclusion in the longest chain.
type MerkleRootsConfirmationsResponse struct {
	AllConfirmed  bool                     `json:"allConfirmed"`
	Confirmations []MerkleRootConfirmation `json:"confirmations"`
}

// newMerkleRootConfirmationcreates a new merkleRootConfirmation
// object from domain's MerkleRootConfirmation object.
func newMerkleRootConfirmation(
	merkleConfms *domains.MerkleRootConfirmation,
) MerkleRootConfirmation {
	return MerkleRootConfirmation{
		Hash:       merkleConfms.Hash,
		MerkleRoot: merkleConfms.MerkleRoot,
		Confirmed:  merkleConfms.Confirmed,
	}
}

// mapToMerkleRootsConfirmationsResponses converts a slice of domain's
// MerkleRootConfirmation objects to merkleRootConfirmationRespose.
func mapToMerkleRootsConfirmationsResponses(
	merkleConfms []*domains.MerkleRootConfirmation,
) MerkleRootsConfirmationsResponse {
	mrcfs := make([]MerkleRootConfirmation, 0)

	allConfirmed := true

	for _, merkleConfm := range merkleConfms {
		mrcfs = append(mrcfs, newMerkleRootConfirmation(merkleConfm))
		if !merkleConfm.Confirmed {
			allConfirmed = false
		}
	}

	return MerkleRootsConfirmationsResponse{
		AllConfirmed:  allConfirmed,
		Confirmations: mrcfs,
	}
}
