package domains

// ExclusiveStartKeyPage is object to use when returning database records in paged format using Exclusive Start Key paging
type ExclusiveStartKeyPage[Content any] struct {
	// List of records for the response
	Content Content `json:"content"`
	// Pagination details
	Page ExclusiveStartKeyPageInfo `json:"page"`
}

// ExclusiveStartKeyPageInfo is object to use when limiting and sorting database query results for Exclusive Start Key Paging
type ExclusiveStartKeyPageInfo struct {
	// Field by which to order the results
	OrderByField *string `json:"orderByField,omitempty"`
	// Direction in which to order the results ASC/DSC
	SortDirection *string `json:"sortDirection,omitempty"`
	// Total count of elements
	TotalElements int32 `json:"totalElements"`
	// Size of the page/returned data
	Size int `json:"size"`
	// Last evaluated key returned from the DB
	LastEvaluatedKey string `json:"lastEvaluatedKey"`
}
