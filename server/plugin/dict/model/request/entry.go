package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type EntrySearch struct {
	DictName string `json:"dictName" form:"dictName"`
	Chinese  string `json:"chinese" form:"chinese"`
	request.PageInfo
}

// TranslateTextReq 单条翻译请求
type TranslateTextReq struct {
	DictName string `json:"dictName" form:"dictName"`
	Chinese  string `json:"chinese" form:"chinese"`
	Lang     string `json:"lang" form:"lang"`
}

// TranslateBatchItem 批量翻译中的单条
type TranslateBatchItem struct {
	Text  string   `json:"text"`
	Langs []string `json:"langs"`
}

// TranslateBatchReq 批量翻译请求
type TranslateBatchReq struct {
	DictName string               `json:"dictName"`
	Items    []TranslateBatchItem `json:"items"`
}
