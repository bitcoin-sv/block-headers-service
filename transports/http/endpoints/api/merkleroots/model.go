package merkleroots

import (
	"github.com/libsv/bitcoin-hc/domains"
)

// merkleRootConfirmationResponse is a confirmation
// of merkle roots inclusion in the longest chain.
type merkleRootConfirmation struct {
	Hash       string `json:"blockhash"`
	MerkleRoot string `json:"merkleRoot"`
	Confirmed  bool   `json:"confirmed"`
}

// merkleRootsConfirmationsResponse is an API reponse for confirming
// merkle roots inclusion in the longest chain.
type merkleRootsConfirmationsResponse struct {
	AllIncluded   bool                     `json:"allIncluded"`
	Confirmations []merkleRootConfirmation `json:"confirmations"`
}

// newMerkleRootConfirmationcreates a new merkleRootConfirmation
// object from domain's MerkleRootConfirmation object.
func newMerkleRootConfirmation(
	merkleConfms *domains.MerkleRootConfirmation,
) merkleRootConfirmation {
	return merkleRootConfirmation{
		Hash:       merkleConfms.Hash,
		MerkleRoot: merkleConfms.MerkleRoot,
		Confirmed:  merkleConfms.Confirmed,
	}
}

// mapToMerkleRootsConfirmationsResponses converts a slice of domain's
// MerkleRootConfirmation objects to merkleRootConfirmationRespose.
func mapToMerkleRootsConfirmationsResponses(
	merkleConfms []*domains.MerkleRootConfirmation,
) merkleRootsConfirmationsResponse {
	mrcfs := make([]merkleRootConfirmation, 0)

	allIncluded := true

	for _, merkleConfm := range merkleConfms {
		mrcfs = append(mrcfs, newMerkleRootConfirmation(merkleConfm))
		if !merkleConfm.Confirmed {
			allIncluded = false
		}
	}

	return merkleRootsConfirmationsResponse{
		AllIncluded:   allIncluded,
		Confirmations: mrcfs,
	}
}
