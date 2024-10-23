package bhserrors

// ////////////////////////////////// AUTH ERRORS

// ErrGeneric is a generic error that something went wrong
var ErrGeneric = BHSError{Message: "Something went wrong. Internal server error", StatusCode: 500, Code: "ErrGeneric"}

// ////////////////////////////////// AUTH ERRORS

// ErrMissingAuthHeader is when request does not have auth header
var ErrMissingAuthHeader = BHSError{Message: "Empty auth header", StatusCode: 401, Code: "ErrMissingAuthHeader"}

// ErrInvalidAuthHeader is when request does not have a valid auth header
var ErrInvalidAuthHeader = BHSError{Message: "Invalid auth header", StatusCode: 401, Code: "ErrInvalidAuthHeader"}

// ErrInvalidAccessToken is when access token is invalid
var ErrInvalidAccessToken = BHSError{Message: "Invalid access token", StatusCode: 401, Code: "ErrInvalidAccessToken"}

// ErrUnauthorized is a generic error when user is unauthorized to make a request
var ErrUnauthorized = BHSError{Message: "Not authorized", StatusCode: 401, Code: "ErrUnauthorized"}

// ////////////////////////////////// MERKLE ROOTS ERRORS

// ErrMerklerootNotFound is when provided merkleroot from user was not found in Block Header Service's database
var ErrMerklerootNotFound = BHSError{Message: "No block with provided merkleroot was found", Code: "ErrMerkleRootNotFound", StatusCode: 404}

// ErrMerklerootNotInLongestChain is when provided merkleroot from user was found in Block Header Service's database but is not in Longest Chain state
var ErrMerklerootNotInLongestChain = BHSError{Message: "Provided merkleroot is not part of the longest chain", Code: "ErrMerkleRootNotInLongestChain", StatusCode: 409}

// ErrInvalidBatchSize is when user provided incorrect batchSize
var ErrInvalidBatchSize = BHSError{Message: "batchSize must be 0 or a positive integer", Code: "ErrInvalidBatchSize", StatusCode: 400}

// ////////////////////////////////// TOKEN ERRORS

// ErrAdminTokenNotFound is when admin token was not found in Block Header Service
var ErrAdminTokenNotFound = BHSError{Message: "Admin token not found", StatusCode: 401, Code: "ErrAdminTokenNotFound"}

// ErrTokenNotFound is when token was not found in Block Header Service
var ErrTokenNotFound = BHSError{Message: "Token not found", StatusCode: 404, Code: "ErrTokenNotFound"}

// ErrCreateToken is when create token fails
var ErrCreateToken = BHSError{Message: "Failed to create new token", StatusCode: 400, Code: "ErrCreateToken"}

// ErrDeleteToken is when delete token fails
var ErrDeleteToken = BHSError{Message: "Failed to delete token", StatusCode: 400, Code: "ErrDeleteToken"}
