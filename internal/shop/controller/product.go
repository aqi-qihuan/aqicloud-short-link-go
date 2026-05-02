package controller

import (
	"strconv"

	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/service"
	"github.com/gin-gonic/gin"
)

type ProductController struct {
	svc *service.ProductService
}

func NewProductController(svc *service.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

// List handles GET /api/product/v1/list.
func (ctrl *ProductController) List(c *gin.Context) {
	products, err := ctrl.svc.List()
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(products))
}

// Detail handles GET /api/product/v1/detail/:product_id.
func (ctrl *ProductController) Detail(c *gin.Context) {
	productId, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		response.JSON(c, response.BuildError("invalid product_id"))
		return
	}
	product, err := ctrl.svc.Detail(productId)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(product))
}
