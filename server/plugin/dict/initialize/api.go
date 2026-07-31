package initialize

import (
	"context"

	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Api(ctx context.Context) {
	entities := []model.SysApi{
		{
			Path:        "/entry/createEntry",
			Description: "新增翻译条目",
			ApiGroup:    "翻译字典",
			Method:      "POST",
		},
		{
			Path:        "/entry/deleteEntry",
			Description: "删除翻译条目",
			ApiGroup:    "翻译字典",
			Method:      "DELETE",
		},
		{
			Path:        "/entry/deleteEntryByIds",
			Description: "批量删除翻译条目",
			ApiGroup:    "翻译字典",
			Method:      "DELETE",
		},
		{
			Path:        "/entry/updateEntry",
			Description: "更新翻译条目",
			ApiGroup:    "翻译字典",
			Method:      "PUT",
		},
		{
			Path:        "/entry/findEntry",
			Description: "根据ID获取翻译条目",
			ApiGroup:    "翻译字典",
			Method:      "GET",
		},
		{
			Path:        "/entry/getEntryList",
			Description: "分页获取翻译条目列表",
			ApiGroup:    "翻译字典",
			Method:      "GET",
		},
		{
			Path:        "/entry/importExcel",
			Description: "上传Excel导入翻译数据",
			ApiGroup:    "翻译字典",
			Method:      "POST",
		},
		{
			Path:        "/entry/loadDictionary",
			Description: "加载完整字典",
			ApiGroup:    "翻译字典",
			Method:      "GET",
		},
		{
			Path:        "/entry/translateText",
			Description: "单条文本翻译",
			ApiGroup:    "翻译字典",
			Method:      "POST",
		},
		{
			Path:        "/entry/translateText",
			Description: "单条文本翻译(GET)",
			ApiGroup:    "翻译字典",
			Method:      "GET",
		},
		{
			Path:        "/entry/translateBatch",
			Description: "批量文本翻译",
			ApiGroup:    "翻译字典",
			Method:      "POST",
		},
	}
	utils.RegisterApis(entities...)
}
