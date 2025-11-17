package types

type SearchQueryParams struct {
	Page int                    `json:"page,omitempty"`
	Size int                    `json:"size,omitempty"`
	Data map[string]interface{} `json:"data,omitempty"`
}
