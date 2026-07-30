package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type Image struct {
	global.GVA_MODEL
	Name string `json:"name" form:"name" gorm:"column:name;comment:图片名称;size:100"`
	Type string `json:"type" form:"type" gorm:"column:type;comment:类型(logo/washLabel);size:50;index"`
	URL  string `json:"url" form:"url" gorm:"column:url;comment:图片地址;size:500"`
	Sort int    `json:"sort" form:"sort" gorm:"column:sort;comment:排序;default:0"`
}

func (Image) TableName() string {
	return "gva_images"
}
