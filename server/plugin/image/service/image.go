package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/image/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/image/model/request"
)

var Image = new(image)

type image struct{}

func (s *image) Create(img *model.Image) error {
	return global.GVA_DB.Create(img).Error
}

func (s *image) Delete(id string) error {
	return global.GVA_DB.Delete(&model.Image{}, "id = ?", id).Error
}

func (s *image) DeleteByIds(ids []string) error {
	return global.GVA_DB.Delete(&[]model.Image{}, "id in ?", ids).Error
}

func (s *image) Update(img model.Image) error {
	return global.GVA_DB.Model(&model.Image{}).Where("id = ?", img.ID).Updates(&img).Error
}

func (s *image) GetByID(id string) (model.Image, error) {
	var img model.Image
	err := global.GVA_DB.Where("id = ?", id).First(&img).Error
	return img, err
}

func (s *image) GetList(info request.ImageSearch) (list []model.Image, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&model.Image{})

	if info.Type != "" {
		db = db.Where("type = ?", info.Type)
	}
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	err = db.Order("sort ASC, id DESC").Find(&list).Error
	return
}
