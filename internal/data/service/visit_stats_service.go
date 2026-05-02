package service

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/aqi/aqicloud-short-link-go/internal/data/vo"
)

type VisitStatsService struct {
	db *sql.DB
}

func NewVisitStatsService(db *sql.DB) *VisitStatsService {
	return &VisitStatsService{db: db}
}

// PageRecord returns paginated visit records.
func (s *VisitStatsService) PageRecord(accountNo int64, code string, page, size int) (*vo.VisitRecordPageVO, error) {
	if page*size > 1000 {
		return nil, fmt.Errorf("query limit exceeded")
	}

	var total int64
	err := s.db.QueryRow("SELECT count(1) FROM visit_stats WHERE account_no = ? AND code = ?", accountNo, code).Scan(&total)
	if err != nil {
		return nil, err
	}

	from := (page - 1) * size
	rows, err := s.db.Query(
		`SELECT code, referer, ts, is_new, account_no, province, city, ip, browser_name, os, device_type, start_time
		 FROM visit_stats WHERE account_no = ? AND code = ?
		 ORDER BY ts DESC LIMIT ?, ?`,
		accountNo, code, from, size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []vo.VisitRecordVO
	for rows.Next() {
		var r vo.VisitRecordVO
		if err := rows.Scan(&r.Code, &r.Referer, &r.VisitTime, &r.IsNew, &r.AccountNo,
			&r.Province, &r.City, &r.IP, &r.BrowserName, &r.OS, &r.DeviceType, &r.StartTime); err != nil {
			return nil, err
		}
		list = append(list, r)
	}

	totalPage := int(math.Ceil(float64(total) / float64(size)))
	return &vo.VisitRecordPageVO{
		Total:       total,
		CurrentPage: page,
		TotalPage:   totalPage,
		Data:        list,
	}, nil
}

// RegionDay returns region distribution for a date range.
func (s *VisitStatsService) RegionDay(accountNo int64, code, startTime, endTime string) ([]vo.RegionDayVO, error) {
	rows, err := s.db.Query(
		`SELECT province, city, sum(pv) AS pv_count, sum(uv) AS uv_count, count(DISTINCT ip) AS ip_count
		 FROM visit_stats WHERE account_no = ? AND code = ?
		 AND toYYYYMMDD(start_time) BETWEEN ? AND ?
		 GROUP BY province, city ORDER BY pv_count DESC`,
		accountNo, code, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []vo.RegionDayVO
	for rows.Next() {
		var r vo.RegionDayVO
		if err := rows.Scan(&r.Province, &r.City, &r.PVCount, &r.UVCount, &r.IPCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// Trend returns visit trend data with DAY/HOUR/MINUTE granularity.
func (s *VisitStatsService) Trend(accountNo int64, code, trendType, startTime, endTime string) ([]vo.VisitTrendVO, error) {
	var query string
	var args []interface{}

	switch trendType {
	case "DAY":
		query = `SELECT toYYYYMMDD(start_time) AS dt,
				 sum(if(is_new = 1, uv, 0)) AS new_uv, sum(uv) AS uv_count,
				 sum(pv) AS pv_count, count(DISTINCT ip) AS ip_count
				 FROM visit_stats WHERE account_no = ? AND code = ?
				 AND toYYYYMMDD(start_time) BETWEEN ? AND ?
				 GROUP BY dt ORDER BY dt ASC`
		args = []interface{}{accountNo, code, startTime, endTime}
	case "HOUR":
		query = `SELECT toString(toHour(start_time)) AS dt,
				 sum(if(is_new = 1, uv, 0)) AS new_uv, sum(uv) AS uv_count,
				 sum(pv) AS pv_count, count(DISTINCT ip) AS ip_count
				 FROM visit_stats WHERE account_no = ? AND code = ?
				 AND toYYYYMMDD(start_time) = ?
				 GROUP BY dt ORDER BY dt ASC`
		args = []interface{}{accountNo, code, startTime}
	case "MINUTE":
		query = `SELECT toString(toMinute(start_time)) AS dt,
				 sum(if(is_new = 1, uv, 0)) AS new_uv, sum(uv) AS uv_count,
				 sum(pv) AS pv_count, count(DISTINCT ip) AS ip_count
				 FROM visit_stats WHERE account_no = ? AND code = ?
				 AND toYYYYMMDDhhmmss(start_time) BETWEEN ? AND ?
				 GROUP BY dt ORDER BY dt ASC`
		args = []interface{}{accountNo, code, startTime, endTime}
	default:
		return nil, fmt.Errorf("unsupported trend type: %s", trendType)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []vo.VisitTrendVO
	for rows.Next() {
		var r vo.VisitTrendVO
		if err := rows.Scan(&r.DateTimeStr, &r.NewUVCount, &r.UVCount, &r.PVCount, &r.IPCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// FrequentIP returns top IPs by visit count.
func (s *VisitStatsService) FrequentIP(accountNo int64, code string) ([]vo.FrequentItemVO, error) {
	return s.frequentQuery(accountNo, code, "ip")
}

// FrequentReferer returns top referrers by visit count.
func (s *VisitStatsService) FrequentReferer(accountNo int64, code string) ([]vo.FrequentItemVO, error) {
	return s.frequentQuery(accountNo, code, "referer")
}

func (s *VisitStatsService) frequentQuery(accountNo int64, code, field string) ([]vo.FrequentItemVO, error) {
	query := fmt.Sprintf(
		`SELECT %s, sum(pv) AS pv_count FROM visit_stats
		 WHERE account_no = ? AND code = ?
		 GROUP BY %s ORDER BY pv_count DESC LIMIT 10`, field, field)

	rows, err := s.db.Query(query, accountNo, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []vo.FrequentItemVO
	for rows.Next() {
		var r vo.FrequentItemVO
		if err := rows.Scan(&r.Name, &r.PVCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// DeviceInfo returns device distribution by field (os/browser/device).
func (s *VisitStatsService) DeviceInfo(accountNo int64, code, field string) ([]vo.DeviceInfoVO, error) {
	var groupCol string
	switch field {
	case "os":
		groupCol = "os"
	case "browser":
		groupCol = "browser_name"
	case "device":
		groupCol = "device_type"
	default:
		return nil, fmt.Errorf("unsupported device field: %s", field)
	}

	query := fmt.Sprintf(
		`SELECT %s, sum(pv) AS pv_count FROM visit_stats
		 WHERE account_no = ? AND code = ?
		 GROUP BY %s ORDER BY pv_count DESC`, groupCol, groupCol)

	rows, err := s.db.Query(query, accountNo, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []vo.DeviceInfoVO
	for rows.Next() {
		var r vo.DeviceInfoVO
		if err := rows.Scan(&r.Name, &r.PVCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}
