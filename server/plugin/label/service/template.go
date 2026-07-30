package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/label/model"
)

var Template = new(template)

type template struct{}

func (s *template) Save(name string, labelWidth int, labelHeight int, headSeam int, marginLR int, needsTranslation bool, dictName string, translateLangs string, elements string, translatedElements string) error {
	var existing model.LabelTemplate
	err := global.GVA_DB.Where("name = ?", name).First(&existing).Error
	if err == nil {
		existing.Elements = elements
		existing.TranslatedElements = translatedElements
		existing.LabelWidth = labelWidth
		existing.LabelHeight = labelHeight
		existing.HeadSeam = headSeam
		existing.MarginLR = marginLR
		existing.NeedsTranslation = needsTranslation
		existing.DictName = dictName
		existing.TranslateLangs = translateLangs
		return global.GVA_DB.Save(&existing).Error
	}
	return global.GVA_DB.Create(&model.LabelTemplate{
		Name:               name,
		LabelWidth:         labelWidth,
		LabelHeight:        labelHeight,
		HeadSeam:           headSeam,
		MarginLR:           marginLR,
		NeedsTranslation:   needsTranslation,
		DictName:           dictName,
		TranslateLangs:     translateLangs,
		Elements:           elements,
		TranslatedElements: translatedElements,
	}).Error
}

func (s *template) Load(name string) (model.LabelTemplate, error) {
	var t model.LabelTemplate
	err := global.GVA_DB.Where("name = ?", name).First(&t).Error
	return t, err
}

func (s *template) List() ([]model.LabelTemplate, error) {
	var list []model.LabelTemplate
	err := global.GVA_DB.Order("id DESC").Find(&list).Error
	return list, err
}

func (s *template) Delete(name string) error {
	return global.GVA_DB.Where("name = ?", name).Delete(&model.LabelTemplate{}).Error
}

// Publish 发布模板
func (s *template) Publish(name string) error {
	return global.GVA_DB.Model(&model.LabelTemplate{}).Where("name = ?", name).Update("is_published", true).Error
}

// Unpublish 取消发布模板
func (s *template) Unpublish(name string) error {
	return global.GVA_DB.Model(&model.LabelTemplate{}).Where("name = ?", name).Update("is_published", false).Error
}

// ListPublished 获取已发布的模板列表
func (s *template) ListPublished() ([]model.LabelTemplate, error) {
	var list []model.LabelTemplate
	err := global.GVA_DB.Where("is_published = ?", true).Order("id DESC").Find(&list).Error
	return list, err
}

// TranslateText 根据字典翻译文本字段
func (s *template) TranslateText(dictName string, chinese string, lang string) (string, error) {
	var entry struct {
		English    string
		Russian    string
		Arabic     string
		Indonesian string
	}
	err := global.GVA_DB.Table("gva_dict_entries").
		Select("english, russian, arabic, indonesian").
		Where("dict_name = ? AND chinese = ?", dictName, chinese).
		First(&entry).Error
	if err != nil {
		return "", err
	}
	switch lang {
	case "english":
		return entry.English, nil
	case "russian":
		return entry.Russian, nil
	case "arabic":
		return entry.Arabic, nil
	case "indonesian":
		return entry.Indonesian, nil
	default:
		return "", nil
	}
}
