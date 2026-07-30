package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type Entry struct {
	global.GVA_MODEL
	DictName   string `json:"dictName" form:"dictName" gorm:"column:dict_name;comment:字典名称(巴拉/森马);size:50;index:idx_dict_chinese,unique"`
	Chinese    string `json:"chinese" form:"chinese" gorm:"column:chinese;comment:中文;size:500;index:idx_dict_chinese,unique"`
	English    string `json:"english" form:"english" gorm:"column:english;comment:英文;size:500"`
	Russian    string `json:"russian" form:"russian" gorm:"column:russian;comment:俄文;size:500"`
	Arabic     string `json:"arabic" form:"arabic" gorm:"column:arabic;comment:阿语译文;size:500"`
	Indonesian string `json:"indonesian" form:"indonesian" gorm:"column:indonesian;comment:印尼语;size:500"`
}

func (Entry) TableName() string {
	return "gva_dict_entries"
}
