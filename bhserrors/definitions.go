package bhserrors

// ////////////////////////////////// MERKLE ROOTS ERRORS

var ErrMerklerootNotFound = BHSError{Message: "No block with provided merkleroot was found", Code: "ErrMerkleRootNotFound", StatusCode: 404}
var ErrMerklerootNotInLongestChain = BHSError{Message: "Provided merkleroot is not part of the longest chain", Code: "ErrMerkleRootNotInLongestChain", StatusCode: 409}
var ErrMerklerootInvalidBatchSize = BHSError{Message: "batchSize must be 0 or a positive integer", Code: "ErrInvalidBatchSize", StatusCode: 400}
