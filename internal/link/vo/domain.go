package vo

// DomainVO is the view object for domain responses.
type DomainVO struct {
	ID         int64  `json:"id"`
	AccountNo  int64  `json:"account_no"`
	DomainType string `json:"domain_type"`
	Value      string `json:"value"`
}
