package bhserrors

// ////////////////////////////////// MERKLE ROOTS ERRORS

// ErrMerklerootNotFound is when provided merkleroot from user was not found in Block Header Service's database
var ErrMerklerootNotFound = BHSError{Message: "No block with provided merkleroot was found", Code: "ErrMerkleRootNotFound", StatusCode: 404}

// ErrMerklerootNotInLongestChain is when provided merkleroot from user was found in Block Header Service's database but is not in Longest Chain state
var ErrMerklerootNotInLongestChain = BHSError{Message: "Provided merkleroot is not part of the longest chain", Code: "ErrMerkleRootNotInLongestChain", StatusCode: 409}

// ErrInvalidBatchSize is when user provided incorrect batchSize
var ErrInvalidBatchSize = BHSError{Message: "batchSize must be 0 or a positive integer", Code: "ErrInvalidBatchSize", StatusCode: 400}
