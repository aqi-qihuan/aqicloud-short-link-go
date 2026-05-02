package vo

// ProductVO is the response for product list/detail.
type ProductVO struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Detail     string  `json:"detail"`
	Img        string  `json:"img"`
	Level      string  `json:"level"`
	OldAmount  float64 `json:"oldAmount"`
	Amount     float64 `json:"amount"`
	PluginType string  `json:"pluginType"`
	DayTimes   int     `json:"dayTimes"`
	TotalTimes int     `json:"totalTimes"`
	ValidDay   int     `json:"validDay"`
}
