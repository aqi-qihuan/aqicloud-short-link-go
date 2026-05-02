package service

import (
	"fmt"
	"log"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"
	accountmodel "github.com/aqi/aqicloud-short-link-go/internal/account/model"
	"github.com/aqi/aqicloud-short-link-go/internal/account/request"
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"gorm.io/gorm"
)

type AccountService struct {
	db    *gorm.DB
	rmq   *mq.RabbitMQ
	notif *NotifyService
}

func NewAccountService(db *gorm.DB, rmq *mq.RabbitMQ, notif *NotifyService) *AccountService {
	return &AccountService{db: db, rmq: rmq, notif: notif}
}

// Register creates a new account and sends free traffic init event.
func (s *AccountService) Register(req *request.AccountRegisterRequest) error {
	// Validate SMS code
	if !s.notif.CheckCode(SendCodeTypeRegister, req.Phone, req.Code) {
		return fmt.Errorf("verification code is incorrect or expired")
	}

	// Check if phone already registered
	var existing accountmodel.AccountDO
	if err := s.db.Where("phone = ?", req.Phone).First(&existing).Error; err == nil {
		return fmt.Errorf("phone number already registered")
	}

	// Generate account number
	accountNo := int64(util.GenerateSnowflakeID())

	// Generate password salt and hash
	salt := "$1$" + util.GetStringNumRandom(8)
	pwdHash := md5CryptHash(req.Pwd, salt)

	account := accountmodel.AccountDO{
		AccountNo: accountNo,
		HeadImg:   req.HeadImg,
		Phone:     req.Phone,
		Pwd:       pwdHash,
		Secret:    salt,
		Mail:      req.Mail,
		Username:  req.Username,
		Auth:      string(enums.AUTH_DEFAULT),
	}

	if err := s.db.Create(&account).Error; err != nil {
		return fmt.Errorf("create account failed: %w", err)
	}

	// Send TRAFFIC_FREE_INIT event to MQ
	if s.rmq != nil {
		eventMsg := model.EventMessage{
			MessageId:        util.GenerateUUID(),
			EventMessageType: string(enums.TRAFFIC_FREE_INIT),
			BizId:            "1", // free product ID
			AccountNo:        accountNo,
		}
		if err := s.rmq.PublishJSON("traffic.event.exchange", "traffic.free_init.routing.key", eventMsg); err != nil {
			log.Printf("[MQ] publish TRAFFIC_FREE_INIT error: %v", err)
		}
	}

	return nil
}

// Login authenticates a user and returns a JWT token.
func (s *AccountService) Login(req *request.AccountLoginRequest) (string, error) {
	var account accountmodel.AccountDO
	if err := s.db.Where("phone = ?", req.Phone).First(&account).Error; err != nil {
		return "", fmt.Errorf("account not found")
	}

	// Re-encrypt password with stored salt and compare
	pwdHash := md5CryptHash(req.Pwd, account.Secret)
	if pwdHash != account.Pwd {
		return "", fmt.Errorf("incorrect password")
	}

	// Build LoginUser and generate JWT token
	loginUser := &model.LoginUser{
		AccountNo: account.AccountNo,
		HeadImg:   account.HeadImg,
		Username:  account.Username,
		Mail:      account.Mail,
		Phone:     account.Phone,
		Auth:      account.Auth,
	}
	token, err := util.GenerateToken(loginUser)
	if err != nil {
		return "", fmt.Errorf("generate token failed: %w", err)
	}

	return token, nil
}

// Detail returns account info by account number.
func (s *AccountService) Detail(accountNo int64) (*accountmodel.AccountDO, error) {
	var account accountmodel.AccountDO
	if err := s.db.Where("account_no = ?", accountNo).First(&account).Error; err != nil {
		return nil, fmt.Errorf("account not found")
	}
	return &account, nil
}

// md5CryptHash encrypts a password with the given salt using MD5-crypt ($1$).
// Compatible with Java's Md5Crypt.md5Crypt(password, salt).
func md5CryptHash(password, salt string) string {
	c := crypt.New(crypt.MD5)
	hash, err := c.Generate([]byte(password), []byte(salt))
	if err != nil {
		log.Printf("[ERROR] md5crypt failed: %v", err)
		return ""
	}
	return hash
}
