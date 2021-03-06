package generic

import (
	"fmt"
	"net/url"
	"strconv"
)

/*
PagingStatistics represent cumulocity's 'application/vnd.com.nsn.cumulocity.pagingStatistics+json'.
See: https://cumulocity.com/guides/reference/rest-implementation/#pagingstatistics-application-vnd-com-nsn-cumulocity-pagingstatistics-json
*/
type PagingStatistics struct {
	TotalRecords int `json:"totalRecords,omitempty"`
	TotalPages   int `json:"totalPages,omitempty"`
	PageSize     int `json:"pageSize"`
	CurrentPage  int `json:"currentPage"`
}

// Appends the query param 'pageSize' to the provided parameter values for a request.
// When provided values is nil an error will be created
func PageSizeParameter(pageSize int, params *url.Values) (error) {
	if pageSize < 1 || pageSize > 2000 {
		return fmt.Errorf("The page size must be between 1 and 2000. Was %d", pageSize)
	}
	if params == nil {
		return fmt.Errorf("The provided parameter values must not be nil!")
	}

	params.Add("pageSize", strconv.Itoa(pageSize))

	return nil
}

