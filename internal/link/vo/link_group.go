package vo

import "time"

// LinkGroupVO is the view object for link group responses.
type LinkGroupVO struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	AccountNo   int64     `json:"account_no"`
	GmtCreate   time.Time `json:"gmt_create"`
	GmtModified time.Time `json:"gmt_modified"`
}
