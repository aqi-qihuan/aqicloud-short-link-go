package vo

import "time"

// ShortLinkVO is the view object for short link responses.
type ShortLinkVO struct {
	ID          int64      `json:"id"`
	GroupID     int64      `json:"group_id"`
	Title       string     `json:"title"`
	OriginalUrl string     `json:"original_url"`
	Domain      string     `json:"domain"`
	Code        string     `json:"code"`
	Sign        string     `json:"sign"`
	Expired     *time.Time `json:"expired"`
	AccountNo   int64      `json:"account_no"`
	State       string     `json:"state"`
	LinkType    string     `json:"link_type"`
	GmtCreate   time.Time  `json:"gmt_create"`
	GmtModified time.Time  `json:"gmt_modified"`
}
