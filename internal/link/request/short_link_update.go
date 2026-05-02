package request

// ShortLinkUpdateRequest is the request body for updating a short link.
type ShortLinkUpdateRequest struct {
	ID          int64  `json:"id"`
	GroupID     int64  `json:"groupId"`
	Title       string `json:"title"`
	OriginalUrl string `json:"originalUrl"`
	Domain      string `json:"domain"`
	DomainType  string `json:"domainType"`
	DomainId    int64  `json:"domainId"`
	Code        string `json:"code"`
	Expired     string `json:"expired"`
}
