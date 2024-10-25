package bhserrors

// ////////////////////////////////// GENERIC ERRORS

// ErrGeneric is a generic error that something went wrong
var ErrGeneric = BHSError{Message: "internal server error", StatusCode: 500, Code: "ErrGeneric"}

// ErrBindBody is an error when it fails to bind JSON body
var ErrBindBody = BHSError{Message: "error during bind JSON body", StatusCode: 400, Code: "ErrBindBody"}

// ////////////////////////////////// AUTH ERRORS

// ErrMissingAuthHeader is when request does not have auth header
var ErrMissingAuthHeader = BHSError{Message: "empty auth header", StatusCode: 401, Code: "ErrMissingAuthHeader"}

// ErrInvalidAuthHeader is when request does not have a valid auth header
var ErrInvalidAuthHeader = BHSError{Message: "invalid auth header", StatusCode: 401, Code: "ErrInvalidAuthHeader"}

// ErrInvalidAccessToken is when access token is invalid
var ErrInvalidAccessToken = BHSError{Message: "invalid access token", StatusCode: 401, Code: "ErrInvalidAccessToken"}

// ErrUnauthorized is a generic error when user is unauthorized to make a request
var ErrUnauthorized = BHSError{Message: "not authorized", StatusCode: 401, Code: "ErrUnauthorized"}

// ErrAdminTokenNotFound is when admin token was not found in Block Header Service
var ErrAdminTokenNotFound = BHSError{Message: "admin token not found", StatusCode: 401, Code: "ErrAdminTokenNotFound"}

// ////////////////////////////////// MERKLE ROOTS ERRORS

// ErrMerklerootNotFound is when provided merkleroot from user was not found in Block Header Service's database
var ErrMerklerootNotFound = BHSError{Message: "no block with provided merkleroot was found", Code: "ErrMerkleRootNotFound", StatusCode: 404}

// ErrMerklerootNotInLongestChain is when provided merkleroot from user was found in Block Header Service's database but is not in Longest Chain state
var ErrMerklerootNotInLongestChain = BHSError{Message: "provided merkleroot is not part of the longest chain", Code: "ErrMerkleRootNotInLongestChain", StatusCode: 409}

// ErrInvalidBatchSize is when user provided incorrect batchSize
var ErrInvalidBatchSize = BHSError{Message: "batchSize must be 0 or a positive integer", Code: "ErrInvalidBatchSize", StatusCode: 400}

// ErrGetChainTipHeight is when it fails to get a chain tip height
var ErrGetChainTipHeight = BHSError{Message: "failed to get chain tip height", Code: "ErrGetChainTipHeight", StatusCode: 400}

// ErrVerifyMerklerootsBadBody is when request for verify merkleroots has wrong body
var ErrVerifyMerklerootsBadBody = BHSError{Message: "at least one merkleroot is required", Code: "ErrVerifyMerklerootsBadBody", StatusCode: 400}

// ////////////////////////////////// ACCESS ERRORS

// ErrTokenNotFound is when token was not found in Block Header Service
var ErrTokenNotFound = BHSError{Message: "token not found", StatusCode: 404, Code: "ErrTokenNotFound"}

// ErrCreateToken is when create token fails
var ErrCreateToken = BHSError{Message: "failed to create new token", StatusCode: 400, Code: "ErrCreateToken"}

// ErrDeleteToken is when delete token fails
var ErrDeleteToken = BHSError{Message: "failed to delete token", StatusCode: 400, Code: "ErrDeleteToken"}

// ////////////////////////////////// HEADERS ERRORS

// ErrAncestorHashHigher is when ancestor hash height is higher than requested header
var ErrAncestorHashHigher = BHSError{Message: "ancestor header height can not be higher than requested header height", StatusCode: 400, Code: "ErrAncestorHashHigher"}

// ErrAncestorNotFound is when ancestor for a given hash was not found
var ErrAncestorNotFound = BHSError{Message: "failed to get ancestor with given hash ", StatusCode: 400, Code: "ErrAncestorNotFound"}

// ErrHeadersNotPartOfTheSameChain is when provided headers are not part of the same chain
var ErrHeadersNotPartOfTheSameChain = BHSError{Message: "the headers provided are not part of the same chain", StatusCode: 400, Code: "ErrHeadersNotPartOfTheSameChain"}

// ErrHeaderWithGivenHashes is when getting header with given hashes fails
var ErrHeaderWithGivenHashes = BHSError{Message: "error during getting headers with given hashes", StatusCode: 400, Code: "ErrHeaderWithGivenHashes"}

// ErrHeaderNotFound is when hash could not be found
var ErrHeaderNotFound = BHSError{Message: "header not found", StatusCode: 404, Code: "ErrHeaderNotFound"}

// ErrHeadersForGivenRangeNotFound is when hash could not be found for given range
var ErrHeadersForGivenRangeNotFound = BHSError{Message: "could not find headers in given range", StatusCode: 404, Code: "ErrHeadersForGivenRangeNotFound"}

// ErrHeaderStopHeightNotFound is when stop height for given heade was not found
var ErrHeaderStopHeightNotFound = BHSError{Message: "could not find stop height for given header", StatusCode: 404, Code: "ErrHeaderStopHeightNotFound"}

// ////////////////////////////////// TIPS ERRORS

// ErrGetTips is when it fails to get tips
var ErrGetTips = BHSError{Message: "failed to get tips", StatusCode: 400, Code: "ErrGetTips"}

// ////////////////////////////////// WEBHOOK ERRORS

// ErrURLBodyRequired is when url is not provided in body
var ErrURLBodyRequired = BHSError{Message: "url is required", StatusCode: 400, Code: "ErrURLBodyRequired"}

// ErrURLParamRequired is when url is not provided in param
var ErrURLParamRequired = BHSError{Message: "url is required", StatusCode: 400, Code: "ErrURLParamRequired"}

// ErrCreateWebhook is when it fails to create a webhook
var ErrCreateWebhook = BHSError{Message: "failed to create a webhook", StatusCode: 400, Code: "ErrCreateWebhook"}

// ErrWebhookNotFound is when webhook was not found
var ErrWebhookNotFound = BHSError{Message: "webhook not found", StatusCode: 404, Code: "ErrWebhookNotFound"}

// ErrRefreshWebhook is when a webhook already exists and is active and we tried to refresh it
var ErrRefreshWebhook = BHSError{Message: "webhook already exists and is active", StatusCode: 400, Code: "ErrRefreshWebhook"}

// ErrGetAllWebhooks is when it failed to get all webhooks
var ErrGetAllWebhooks = BHSError{Message: "failed to get all webhooks", StatusCode: 400, Code: "ErrGetAllWebhooks"}

// ErrDeleteWebhook is when it failed to delete a webhook
var ErrDeleteWebhook = BHSError{Message: "failed to delete webhook", StatusCode: 400, Code: "ErrDeleteWebhook"}
