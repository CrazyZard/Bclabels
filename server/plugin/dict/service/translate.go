package service

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/model"
)

// LoadDict 从数据库加载指定字典到内存
func LoadDict(dictName string) (*Dict, error) {
	var entries []model.Entry
	if err := global.GVA_DB.Where("dict_name = ?", dictName).Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("字典「%s」为空，请先在翻译字典中导入词条", dictName)
	}
	d := &Dict{Entries: make(map[string]map[string]string, len(entries))}
	for _, e := range entries {
		m := map[string]string{}
		if e.English != "" {
			m["english"] = e.English
		}
		if e.Russian != "" {
			m["russian"] = e.Russian
		}
		if e.Arabic != "" {
			m["arabic"] = e.Arabic
		}
		if e.Indonesian != "" {
			m["indonesian"] = e.Indonesian
		}
		d.Entries[e.Chinese] = m
	}
	return d, nil
}

// TranslateResult 单条文本的多语言译文
type TranslateResult struct {
	Text         string            `json:"text"`
	Translations map[string]string `json:"translations"`
}

// TranslateText 使用完整翻译引擎翻译单条文本
func (s *entry) TranslateText(dictName, chinese, lang string) (string, error) {
	d, err := LoadDict(dictName)
	if err != nil {
		return "", err
	}
	return d.TranslateText(chinese, lang), nil
}

// TranslateBatch 批量翻译：每条中文对应一组目标语言
func (s *entry) TranslateBatch(dictName string, items []TranslateBatchItem) ([]TranslateResult, []string, error) {
	d, err := LoadDict(dictName)
	if err != nil {
		return nil, nil, err
	}
	results := make([]TranslateResult, len(items))
	for i, item := range items {
		tr := TranslateResult{
			Text:         item.Text,
			Translations: make(map[string]string, len(item.Langs)),
		}
		for _, lang := range item.Langs {
			if lang == "" {
				continue
			}
			tr.Translations[lang] = d.TranslateText(item.Text, lang)
		}
		results[i] = tr
	}
	return results, d.MissList(), nil
}

// TranslateBatchItem 批量翻译请求项
type TranslateBatchItem struct {
	Text  string   `json:"text"`
	Langs []string `json:"langs"`
}
