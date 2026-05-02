package service

import (
	shopmodel "github.com/aqi/aqicloud-short-link-go/internal/shop/model"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/vo"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// List returns all products.
func (s *ProductService) List() ([]vo.ProductVO, error) {
	var products []shopmodel.ProductDO
	if err := s.db.Order("id ASC").Find(&products).Error; err != nil {
		return nil, err
	}
	result := make([]vo.ProductVO, len(products))
	for i, p := range products {
		result[i] = toProductVO(p)
	}
	return result, nil
}

// Detail returns a single product by ID.
func (s *ProductService) Detail(productId int64) (*vo.ProductVO, error) {
	var product shopmodel.ProductDO
	if err := s.db.Where("id = ?", productId).First(&product).Error; err != nil {
		return nil, err
	}
	vo := toProductVO(product)
	return &vo, nil
}

func toProductVO(p shopmodel.ProductDO) vo.ProductVO {
	return vo.ProductVO{
		ID:         p.ID,
		Title:      p.Title,
		Detail:     p.Detail,
		Img:        p.Img,
		Level:      p.Level,
		OldAmount:  p.OldAmount,
		Amount:     p.Amount,
		PluginType: p.PluginType,
		DayTimes:   p.DayTimes,
		TotalTimes: p.TotalTimes,
		ValidDay:   p.ValidDay,
	}
}
