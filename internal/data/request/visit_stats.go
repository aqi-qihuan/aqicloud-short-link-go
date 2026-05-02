package request

// VisitRecordPageRequest is the body for POST /api/visit_stats/v1/page_record.
type VisitRecordPageRequest struct {
	Code  string `json:"code"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

// RegionQueryRequest is the body for POST /api/visit_stats/v1/region_day.
type RegionQueryRequest struct {
	Code      string `json:"code"`
	StartTime string `json:"startTime"` // YYYYMMDD
	EndTime   string `json:"endTime"`   // YYYYMMDD
}

// VisitTrendQueryRequest is the body for POST /api/visit_stats/v1/trend.
type VisitTrendQueryRequest struct {
	Code      string `json:"code"`
	Type      string `json:"type"`      // DAY, HOUR, MINUTE
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// FrequentRequest is the body for POST /api/visit_stats/v1/frequent_ip and /frequent_referer.
type FrequentRequest struct {
	Code      string `json:"code"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// DeviceInfoRequest is the body for POST /api/visit_stats/v1/device_info.
type DeviceInfoRequest struct {
	Code      string `json:"code"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Field     string `json:"field"` // os, browser, device
}
