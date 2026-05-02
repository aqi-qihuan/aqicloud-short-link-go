package vo

// VisitStatsVO is the response for visit statistics queries.
type VisitStatsVO struct {
	Code        string `json:"code,omitempty"`
	Referer     string `json:"referer,omitempty"`
	IsNew       int    `json:"isNew,omitempty"`
	AccountNo   int64  `json:"accountNo,omitempty"`
	VisitTime   int64  `json:"visitTime,omitempty"`
	Province    string `json:"province,omitempty"`
	City        string `json:"city,omitempty"`
	IP          string `json:"ip,omitempty"`
	BrowserName string `json:"browserName,omitempty"`
	OS          string `json:"os,omitempty"`
	DeviceType  string `json:"deviceType,omitempty"`
	StartTime   string `json:"startTime,omitempty"`
	EndTime     string `json:"endTime,omitempty"`
	PV          int64  `json:"pv,omitempty"`
	UV          int64  `json:"uv,omitempty"`
	PVCount     int64  `json:"pvCount,omitempty"`
	UVCount     int64  `json:"uvCount,omitempty"`
	IPCount     int64  `json:"ipCount,omitempty"`
	NewUVCount  int64  `json:"newUVCount,omitempty"`
	DateTimeStr string `json:"dateTimeStr,omitempty"`
}

// VisitRecordVO represents a single visit record.
type VisitRecordVO struct {
	Code        string `json:"code"`
	Referer     string `json:"referer"`
	VisitTime   int64  `json:"visitTime"`
	IsNew       int    `json:"isNew"`
	AccountNo   int64  `json:"accountNo"`
	Province    string `json:"province"`
	City        string `json:"city"`
	IP          string `json:"ip"`
	BrowserName string `json:"browserName"`
	OS          string `json:"os"`
	DeviceType  string `json:"deviceType"`
	StartTime   string `json:"startTime"`
}

// VisitRecordPageVO is the paginated response for visit records.
type VisitRecordPageVO struct {
	Total       int64           `json:"total"`
	CurrentPage int             `json:"current_page"`
	TotalPage   int             `json:"total_page"`
	Data        []VisitRecordVO `json:"data"`
}

// RegionDayVO represents region distribution for a day.
type RegionDayVO struct {
	Province string `json:"province"`
	City     string `json:"city"`
	PVCount  int64  `json:"pvCount"`
	UVCount  int64  `json:"uvCount"`
	IPCount  int64  `json:"ipCount"`
}

// VisitTrendVO represents a single trend data point.
type VisitTrendVO struct {
	DateTimeStr string `json:"dateTimeStr"`
	NewUVCount  int64  `json:"newUVCount"`
	UVCount     int64  `json:"uvCount"`
	PVCount     int64  `json:"pvCount"`
	IPCount     int64  `json:"ipCount"`
}

// FrequentItemVO represents a frequent IP/referer item.
type FrequentItemVO struct {
	Name    string `json:"name"`
	PVCount int64  `json:"pvCount"`
}

// DeviceInfoVO represents a device info distribution item.
type DeviceInfoVO struct {
	Name    string `json:"name"`
	PVCount int64  `json:"pvCount"`
}
