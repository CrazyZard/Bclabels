package initialize

import (
	"context"

	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/label/saveTemplate", Description: "保存标签模板", ApiGroup: "标签管理", Method: "POST"},
		{Path: "/label/deleteTemplate", Description: "删除标签模板", ApiGroup: "标签管理", Method: "DELETE"},
		{Path: "/label/publishTemplate", Description: "发布模板", ApiGroup: "标签管理", Method: "POST"},
		{Path: "/label/unpublishTemplate", Description: "取消发布模板", ApiGroup: "标签管理", Method: "POST"},
		{Path: "/label/loadTemplate", Description: "加载标签模板", ApiGroup: "标签管理", Method: "GET"},
		{Path: "/label/listTemplate", Description: "标签模板列表", ApiGroup: "标签管理", Method: "GET"},
		{Path: "/label/listPublishedTemplate", Description: "已发布模板列表", ApiGroup: "标签管理", Method: "GET"},
		{Path: "/label/translateText", Description: "翻译文本", ApiGroup: "标签管理", Method: "GET"},
		{Path: "/label/downloadBatchTemplate", Description: "下载批量导入Excel模板", ApiGroup: "标签管理", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
