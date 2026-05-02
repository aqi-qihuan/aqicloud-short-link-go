package request

// ShortLinkAddRequest is the request body for creating a short link.
type ShortLinkAddRequest struct {
	GroupID     int64  `json:"groupId"`
	Title       string `json:"title"`
	OriginalUrl string `json:"originalUrl"`
	DomainID    int64  `json:"domainId"`
	DomainType  string `json:"domainType"`
	Expired     string `json:"expired"`
}
