package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/aqi/aqicloud-short-link-go/internal/link/component"
	linkmodel "github.com/aqi/aqicloud-short-link-go/internal/link/model"
	"github.com/aqi/aqicloud-short-link-go/internal/link/request"
	"github.com/aqi/aqicloud-short-link-go/internal/link/sharding"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ShortLinkService struct {
	dbs         []*gorm.DB // ds0, ds1, dsa
	rdb         *redis.Client
	accountAddr string // e.g. "http://localhost:8001"
	rpcToken    string // inter-service auth token
}

func NewShortLinkService(dbs []*gorm.DB, rdb *redis.Client, accountAddr, rpcToken string) *ShortLinkService {
	return &ShortLinkService{dbs: dbs, rdb: rdb, accountAddr: accountAddr, rpcToken: rpcToken}
}

// HandleAddShortLink processes add events from both link and mapping queues.
// Implements recursive collision retry matching Java behavior.
func (s *ShortLinkService) HandleAddShortLink(eventMessage *model.EventMessage) bool {
	var addReq request.ShortLinkAddRequest
	if err := json.Unmarshal([]byte(eventMessage.Content), &addReq); err != nil {
		log.Printf("[MQ] unmarshal add request failed: %v", err)
		return false
	}

	accountNo := eventMessage.AccountNo
	messageType := eventMessage.EventMessageType

	// Generate short link code
	shortLinkCode := component.CreateShortLinkCode(addReq.OriginalUrl)

	// Acquire distributed lock
	acquired := s.acquireCodeLock(shortLinkCode, accountNo)
	if !acquired {
		// Lock failed, wait 100ms then retry with versioned URL
		time.Sleep(100 * time.Millisecond)
		return s.retryAddWithVersion(eventMessage, &addReq)
	}

	if messageType == string(enums.SHORT_LINK_ADD_LINK) {
		return s.handleAddCLink(eventMessage, &addReq, shortLinkCode, accountNo)
	} else if messageType == string(enums.SHORT_LINK_ADD_MAPPING) {
		return s.handleAddMapping(eventMessage, &addReq, shortLinkCode, accountNo)
	}
	return false
}

// handleAddCLink inserts into short_link table (C-side / client-facing).
func (s *ShortLinkService) handleAddCLink(eventMessage *model.EventMessage, addReq *request.ShortLinkAddRequest, code string, accountNo int64) bool {
	dbPrefix, tableSuffix := sharding.RouteShortLink(code)
	dbIdx := sharding.GetDBIndexByPrefix(dbPrefix)
	tableName := sharding.GetTableName("short_link", tableSuffix)

	var existing linkmodel.ShortLinkDO
	err := s.dbs[dbIdx].Table(tableName).Where("code = ? AND del = 0", code).First(&existing).Error
	if err == nil {
		return s.retryAddWithVersion(eventMessage, addReq)
	}
	if err != gorm.ErrRecordNotFound {
		log.Printf("[MQ] check short_link code error: %v", err)
		return false
	}

	// Deduct traffic (call account service)
	if !s.reduceTraffic(accountNo, code) {
		log.Printf("[MQ] traffic reduce failed for account %d", accountNo)
		return false
	}

	sign := util.MD5(addReq.OriginalUrl)
	shortLinkDO := linkmodel.ShortLinkDO{
		ID:          int64(util.GenerateSnowflakeID()),
		AccountNo:   accountNo,
		Code:        code,
		Title:       addReq.Title,
		OriginalUrl: addReq.OriginalUrl,
		Domain:      addReq.DomainType,
		GroupID:     addReq.GroupID,
		Expired:     parseExpired(addReq.Expired),
		Sign:        sign,
		State:       string(enums.SL_ACTIVE),
		Del:         0,
	}

	if err := s.dbs[dbIdx].Table(tableName).Create(&shortLinkDO).Error; err != nil {
		log.Printf("[MQ] insert short_link error: %v", err)
		return false
	}
	return true
}

// handleAddMapping inserts into group_code_mapping table (B-side / admin).
func (s *ShortLinkService) handleAddMapping(eventMessage *model.EventMessage, addReq *request.ShortLinkAddRequest, code string, accountNo int64) bool {
	dbIdx, tableIdx := sharding.RouteGroupCodeMapping(accountNo, addReq.GroupID)
	tableName := sharding.GetTableName("group_code_mapping", fmt.Sprintf("%d", tableIdx))

	var existing linkmodel.GroupCodeMappingDO
	err := s.dbs[dbIdx].Table(tableName).
		Where("code = ? AND account_no = ? AND del = 0 AND group_id = ?", code, accountNo, addReq.GroupID).
		First(&existing).Error
	if err == nil {
		return s.retryAddWithVersion(eventMessage, addReq)
	}
	if err != gorm.ErrRecordNotFound {
		log.Printf("[MQ] check group_code_mapping error: %v", err)
		return false
	}

	sign := util.MD5(addReq.OriginalUrl)
	mappingDO := linkmodel.GroupCodeMappingDO{
		ID:          int64(util.GenerateSnowflakeID()),
		AccountNo:   accountNo,
		Code:        code,
		Title:       addReq.Title,
		OriginalUrl: addReq.OriginalUrl,
		Domain:      addReq.DomainType,
		GroupID:     addReq.GroupID,
		Expired:     parseExpired(addReq.Expired),
		Sign:        sign,
		State:       string(enums.SL_ACTIVE),
		Del:         0,
	}

	if err := s.dbs[dbIdx].Table(tableName).Create(&mappingDO).Error; err != nil {
		log.Printf("[MQ] insert group_code_mapping error: %v", err)
		return false
	}
	return true
}

// HandleDelShortLink processes delete events (soft delete).
func (s *ShortLinkService) HandleDelShortLink(eventMessage *model.EventMessage) bool {
	var delReq request.ShortLinkDelRequest
	if err := json.Unmarshal([]byte(eventMessage.Content), &delReq); err != nil {
		log.Printf("[MQ] unmarshal del request failed: %v", err)
		return false
	}

	accountNo := eventMessage.AccountNo
	messageType := eventMessage.EventMessageType

	if messageType == string(enums.SHORT_LINK_DEL_LINK) {
		dbPrefix, tableSuffix := sharding.RouteShortLink(delReq.Code)
		dbIdx := sharding.GetDBIndexByPrefix(dbPrefix)
		tableName := sharding.GetTableName("short_link", tableSuffix)

		result := s.dbs[dbIdx].Table(tableName).
			Where("code = ? AND account_no = ?", delReq.Code, accountNo).
			Update("del", 1)
		if result.Error != nil {
			log.Printf("[MQ] del short_link error: %v", result.Error)
			return false
		}
		log.Printf("[MQ] deleted C-side short link, rows=%d", result.RowsAffected)
		return true

	} else if messageType == string(enums.SHORT_LINK_DEL_MAPPING) {
		dbIdx, tableIdx := sharding.RouteGroupCodeMapping(accountNo, delReq.GroupID)
		tableName := sharding.GetTableName("group_code_mapping", fmt.Sprintf("%d", tableIdx))

		result := s.dbs[dbIdx].Table(tableName).
			Where("id = ? AND account_no = ? AND group_id = ?", delReq.MappingID, accountNo, delReq.GroupID).
			Update("del", 1)
		if result.Error != nil {
			log.Printf("[MQ] del group_code_mapping error: %v", result.Error)
			return false
		}
		log.Printf("[MQ] deleted B-side mapping, rows=%d", result.RowsAffected)
		return true
	}
	return false
}

// HandleUpdateShortLink processes update events.
func (s *ShortLinkService) HandleUpdateShortLink(eventMessage *model.EventMessage) bool {
	var updateReq request.ShortLinkUpdateRequest
	if err := json.Unmarshal([]byte(eventMessage.Content), &updateReq); err != nil {
		log.Printf("[MQ] unmarshal update request failed: %v", err)
		return false
	}

	accountNo := eventMessage.AccountNo
	messageType := eventMessage.EventMessageType

	if messageType == string(enums.SHORT_LINK_UPDATE_LINK) {
		dbPrefix, tableSuffix := sharding.RouteShortLink(updateReq.Code)
		dbIdx := sharding.GetDBIndexByPrefix(dbPrefix)
		tableName := sharding.GetTableName("short_link", tableSuffix)

		result := s.dbs[dbIdx].Table(tableName).
			Where("code = ? AND account_no = ? AND del = 0", updateReq.Code, accountNo).
			Updates(map[string]interface{}{
				"title":  updateReq.Title,
				"domain": updateReq.Domain,
			})
		if result.Error != nil {
			log.Printf("[MQ] update short_link error: %v", result.Error)
			return false
		}
		log.Printf("[MQ] updated C-side short link, rows=%d", result.RowsAffected)
		return true

	} else if messageType == string(enums.SHORT_LINK_UPDATE_MAPPING) {
		dbIdx, tableIdx := sharding.RouteGroupCodeMapping(accountNo, updateReq.GroupID)
		tableName := sharding.GetTableName("group_code_mapping", fmt.Sprintf("%d", tableIdx))

		result := s.dbs[dbIdx].Table(tableName).
			Where("id = ? AND account_no = ? AND group_id = ? AND del = 0", updateReq.ID, accountNo, updateReq.GroupID).
			Updates(map[string]interface{}{
				"title":  updateReq.Title,
				"domain": updateReq.Domain,
			})
		if result.Error != nil {
			log.Printf("[MQ] update group_code_mapping error: %v", result.Error)
			return false
		}
		log.Printf("[MQ] updated B-side mapping, rows=%d", result.RowsAffected)
		return true
	}
	return false
}

// retryAddWithVersion increments URL version and retries.
func (s *ShortLinkService) retryAddWithVersion(eventMessage *model.EventMessage, addReq *request.ShortLinkAddRequest) bool {
	newUrl := util.AddUrlPrefixVersion(addReq.OriginalUrl)
	addReq.OriginalUrl = newUrl
	content, _ := json.Marshal(addReq)
	eventMessage.Content = string(content)
	return s.HandleAddShortLink(eventMessage)
}

// acquireCodeLock tries to acquire a distributed lock via Redis Lua.
func (s *ShortLinkService) acquireCodeLock(code string, accountNo int64) bool {
	if s.rdb == nil {
		return true
	}
	script := redis.NewScript(`
		if redis.call('EXISTS',KEYS[1])==0 then
			redis.call('set',KEYS[1],ARGV[1]);
			redis.call('expire',KEYS[1],ARGV[2]);
			return 1;
		elseif redis.call('get',KEYS[1]) == ARGV[1] then
			return 2;
		else
			return 0;
		end
	`)
	result, err := script.Run(context.Background(), s.rdb, []string{code},
		fmt.Sprintf("%d", accountNo), 100).Int()
	if err != nil {
		log.Printf("[Redis] acquire code lock error: %v", err)
		return false
	}
	return result == 1 || result == 2
}

// reduceTraffic calls the account service to deduct traffic via HTTP RPC.
func (s *ShortLinkService) reduceTraffic(accountNo int64, bizId string) bool {
	if s.accountAddr == "" {
		return true // skip if not configured
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"accountNo": accountNo,
		"bizId":     bizId,
	})
	req, err := http.NewRequest("POST", s.accountAddr+"/api/traffic/v1/reduce", bytes.NewReader(payload))
	if err != nil {
		log.Printf("[RPC] create traffic reduce request error: %v", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("rpc-token", s.rpcToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[RPC] traffic reduce call error: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Code == 0
}

func parseExpired(expired string) *time.Time {
	if expired == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", expired)
	if err != nil {
		t, err = time.Parse("2006-01-02", expired)
		if err != nil {
			return nil
		}
	}
	return &t
}
