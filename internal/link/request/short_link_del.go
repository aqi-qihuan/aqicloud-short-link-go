package request

// ShortLinkDelRequest is the request body for deleting a short link.
type ShortLinkDelRequest struct {
	GroupID   int64  `json:"groupId"`
	MappingID int64  `json:"mappingId"`
	Code      string `json:"code"`
}
