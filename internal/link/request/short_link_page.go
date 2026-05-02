package request

// ShortLinkPageRequest is the request body for paging short links.
type ShortLinkPageRequest struct {
	Page    int `json:"page"`
	Size    int `json:"size"`
	GroupID int64 `json:"groupId"`
}
