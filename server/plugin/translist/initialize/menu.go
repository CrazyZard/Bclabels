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
			Path:      "translist",
			Name:      "translist",
			Hidden:    false,
			Component: "plugin/translist/view/index.vue",
			Sort:      9,
			Meta:      model.Meta{Title: "翻译列表", Icon: "document"},
		},
	}
	utils.RegisterMenus(entities...)
}
