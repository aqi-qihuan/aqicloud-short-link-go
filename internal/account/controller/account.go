package controller

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aqi/aqicloud-short-link-go/internal/account/request"
	"github.com/aqi/aqicloud-short-link-go/internal/account/service"
	"github.com/aqi/aqicloud-short-link-go/internal/account/vo"
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/common/storage"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/gin-gonic/gin"
)

type AccountController struct {
	svc     *service.AccountService
	storage storage.Storage
}

func NewAccountController(svc *service.AccountService, store storage.Storage) *AccountController {
	return &AccountController{svc: svc, storage: store}
}

// Register handles POST /api/account/v1/register.
func (ctrl *AccountController) Register(c *gin.Context) {
	var req request.AccountRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}
	if req.Phone == "" || req.Pwd == "" || req.Code == "" {
		response.JSON(c, response.BuildError("phone, pwd, code are required"))
		return
	}

	if err := ctrl.svc.Register(&req); err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}

	response.JSON(c, response.BuildSuccess())
}

// Login handles POST /api/account/v1/login.
func (ctrl *AccountController) Login(c *gin.Context) {
	var req request.AccountLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}
	if req.Phone == "" || req.Pwd == "" {
		response.JSON(c, response.BuildError("phone and pwd are required"))
		return
	}

	token, err := ctrl.svc.Login(&req)
	if err != nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}

	response.JSON(c, response.BuildSuccessData(token))
}

// Detail handles GET /api/account/v1/detail.
func (ctrl *AccountController) Detail(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}

	account, err := ctrl.svc.Detail(loginUser.AccountNo)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}

	accountVO := vo.AccountVO{
		AccountNo: account.AccountNo,
		HeadImg:   account.HeadImg,
		Phone:     account.Phone,
		Mail:      account.Mail,
		Username:  account.Username,
		Auth:      account.Auth,
		GmtCreate: account.GmtCreate,
	}

	response.JSON(c, response.BuildSuccessData(accountVO))
}

// Upload handles POST /api/account/v1/upload (multipart file upload).
func (ctrl *AccountController) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.JSON(c, response.BuildError("file is required"))
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]string{
		".jpg": "image/jpeg", ".jpeg": "image/jpeg",
		".png": "image/png", ".gif": "image/gif",
	}
	contentType, ok := allowed[ext]
	if !ok {
		response.JSON(c, response.BuildError("unsupported file type, allowed: jpg/jpeg/png/gif"))
		return
	}

	// Read file content for hashing
	data, err := io.ReadAll(file)
	if err != nil {
		response.JSON(c, response.BuildError("read file failed"))
		return
	}

	// Limit file size to 5MB
	if len(data) > 5*1024*1024 {
		response.JSON(c, response.BuildError("file size exceeds 5MB limit"))
		return
	}

	// Generate unique object key
	hash := fmt.Sprintf("%X", md5.Sum(data))
	objectKey := storage.GenerateObjectKey(header.Filename, hash)

	// Upload to storage
	url, err := ctrl.storage.Upload(objectKey, io.NopCloser(strings.NewReader(string(data))), contentType)
	if err != nil {
		response.JSON(c, response.BuildError("upload failed: "+err.Error()))
		return
	}

	response.JSON(c, response.BuildSuccessData(gin.H{
		"url":      url,
		"filename": header.Filename,
		"size":     len(data),
	}))
}

// Captcha handles GET /api/account/v1/captcha (returns JPEG image).
// NOTE: Apifox spec shows this under /api/account/v1, not /api/notify/v1.
func (ctrl *AccountController) Captcha(svc *service.NotifyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		captcha := svc.GenerateCaptcha()

		// Key = MD5(ip + userAgent)
		key := util.MD5(c.ClientIP() + c.GetHeader("User-Agent"))
		if err := svc.SaveCaptcha(key, captcha); err != nil {
			c.String(http.StatusInternalServerError, "save captcha failed")
			return
		}

		imgBytes, err := util.CaptchaImage(captcha)
		if err != nil {
			c.String(http.StatusInternalServerError, "generate captcha image failed")
			return
		}

		c.Data(http.StatusOK, "image/jpeg", imgBytes)
	}
}

// SendCode handles POST /api/account/v1/send_code.
func (ctrl *AccountController) SendCode(svc *service.NotifyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request.SendCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.JSON(c, response.BuildError("invalid request body"))
			return
		}
		if req.Captcha == "" || req.To == "" {
			response.JSON(c, response.BuildError("captcha and to are required"))
			return
		}

		// Validate captcha
		key := util.MD5(c.ClientIP() + c.GetHeader("User-Agent"))
		if !svc.ValidateCaptcha(key, req.Captcha) {
			response.JSON(c, response.BuildError("captcha is incorrect or expired"))
			return
		}

		if err := svc.SendCode(service.SendCodeTypeRegister, req.To); err != nil {
			response.JSON(c, response.BuildError(err.Error()))
			return
		}

		response.JSON(c, response.BuildSuccess())
	}
}

