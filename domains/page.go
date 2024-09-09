package domains

// ExclusiveStartKeyPagedResponse is object to use when returning database records in paged format using Exclusive Start Key paging
type ExclusiveStartKeyPagedResponse[content any] struct {
	// List of records for the response
	Content []content `json:"content"`
	// Pagination details
	Page ExclusiveStartKeyPage `json:"page"`
}

// ExclusiveStartKeyPage is object to use when limiting and sorting database query results for Exclusive Start Key Paging
type ExclusiveStartKeyPage struct {
	// Field by which to order the results
	OrderByField *string `json:"orderByField"`
	// Direction in which to order the results ASC/DSC
	SortDirection *string `json:"sortDirection"`
	// Total count of elements
	TotalElements int32 `json:"totalElements"`
	// Size of the page/returned data
	Size int `json:"size"`
	// Last evaluated key returned from the DB
	LastEvaluatedKey any `json:"lastEvaluatedKey"`
}
