package merkleroots

import (
	"github.com/libsv/bitcoin-hc/domains"
)

// merkleRootConfirmationRespose is an API response for
// confirmation of merkle roots inclusion in the longest chain.
type merkleRootConfirmationRespose struct {
	Hash       string `json:"blockhash"`
	MerkleRoot string `json:"merkleRoot"`
	Confirmed  bool   `json:"confirmed"`
}

// newMerkleRootConfirmationReponse creates a new merkleRootConfirmationRespose
// object from domain's MerkleRootConfirmation object.
func newMerkleRootConfirmationReponse(
	merkleConfms *domains.MerkleRootConfirmation,
) merkleRootConfirmationRespose {
	return merkleRootConfirmationRespose{
		Hash:       merkleConfms.Hash,
		MerkleRoot: merkleConfms.MerkleRoot,
		Confirmed:  merkleConfms.Confirmed,
	}
}

// mapToMerkleRootsConfirmationsResponses converts a slice of domain's
// MerkleRootConfirmation objects to merkleRootConfirmationRespose.
func mapToMerkleRootsConfirmationsResponses(
	merkleConfms []*domains.MerkleRootConfirmation,
) []merkleRootConfirmationRespose {
	merkleConfmsResps := make([]merkleRootConfirmationRespose, 0)

	for _, merkleConfm := range merkleConfms {
		merkleConfmsResps = append(
			merkleConfmsResps,
			newMerkleRootConfirmationReponse(merkleConfm),
		)
	}

	return merkleConfmsResps
}
