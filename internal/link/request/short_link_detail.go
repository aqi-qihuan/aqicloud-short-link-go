package request

// ShortLinkDetailRequest is the request body for getting short link detail.
type ShortLinkDetailRequest struct {
	GroupID   int64 `json:"groupId"`
	MappingID int64 `json:"mappingId"`
}
