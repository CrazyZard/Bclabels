package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type LabelTemplate struct {
	global.GVA_MODEL
	Name               string `json:"name" gorm:"column:name;comment:模板名称;size:100;uniqueIndex"`
	LabelWidth         int    `json:"labelWidth" gorm:"column:label_width;comment:标签宽度(mm);default:80"`
	LabelHeight        int    `json:"labelHeight" gorm:"column:label_height;comment:标签高度(mm);default:120"`
	HeadSeam           int    `json:"headSeam" gorm:"column:head_seam;comment:顶缝(mm);default:8"`
	MarginLR           int    `json:"marginLR" gorm:"column:margin_lr;comment:左右空白间距(mm);default:2"`
	NeedsTranslation   bool   `json:"needsTranslation" gorm:"column:needs_translation;comment:需要翻译面;default:false"`
	DictName           string `json:"dictName" gorm:"column:dict_name;comment:翻译字典名称;size:50"`
	TranslateLangs     string `json:"translateLangs" gorm:"column:translate_langs;comment:翻译语言列表JSON;size:200"`
	Elements           string `json:"elements" gorm:"column:elements;comment:正面元素JSON;type:longtext"`
	TranslatedElements string `json:"translatedElements" gorm:"column:translated_elements;comment:反面元素JSON;type:longtext"`
	IsPublished        bool   `json:"isPublished" gorm:"column:is_published;comment:是否已发布;default:false"`
}

func (LabelTemplate) TableName() string {
	return "gva_label_templates"
}
