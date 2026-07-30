package initialize

import (
	"context"

	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Api(ctx context.Context) {
	entities := []model.SysApi{
		{
			Path:        "/translist/uploadAndTranslate",
			Description: "上传Excel并翻译",
			ApiGroup:    "翻译列表",
			Method:      "POST",
		},
		{
			Path:        "/translist/deleteJob",
			Description: "删除翻译任务",
			ApiGroup:    "翻译列表",
			Method:      "DELETE",
		},
		{
			Path:        "/translist/deleteJobByIds",
			Description: "批量删除翻译任务",
			ApiGroup:    "翻译列表",
			Method:      "DELETE",
		},
		{
			Path:        "/translist/findJob",
			Description: "查询翻译任务",
			ApiGroup:    "翻译列表",
			Method:      "GET",
		},
		{
			Path:        "/translist/getJobList",
			Description: "分页获取翻译任务列表",
			ApiGroup:    "翻译列表",
			Method:      "GET",
		},
		{
			Path:        "/translist/retranslate",
			Description: "重新翻译",
			ApiGroup:    "翻译列表",
			Method:      "POST",
		},
		{
			Path:        "/translist/exportExcel",
			Description: "导出翻译结果Excel",
			ApiGroup:    "翻译列表",
			Method:      "GET",
		},
	}
	utils.RegisterApis(entities...)
}
