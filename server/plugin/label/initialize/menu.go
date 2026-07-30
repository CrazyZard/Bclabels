package initialize

import (
	"context"

	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			ParentId:  9,
			Path:      "labelEditor",
			Name:      "labelEditor",
			Hidden:    false,
			Component: "plugin/label/view/index.vue",
			Sort:      7,
			Meta:      model.Meta{Title: "标签管理", Icon: "tickets"},
		},
		{
			ParentId:  9,
			Path:      "batchLabelEditor",
			Name:      "batchLabelEditor",
			Hidden:    false,
			Component: "plugin/label/view/batch.vue",
			Sort:      8,
			Meta:      model.Meta{Title: "生产模板", Icon: "printer"},
		},
	}
	utils.RegisterMenus(entities...)
}
