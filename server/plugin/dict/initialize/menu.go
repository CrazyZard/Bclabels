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
			Path:      "dictEntry",
			Name:      "dictEntry",
			Hidden:    false,
			Component: "plugin/dict/view/index.vue",
			Sort:      6,
			Meta:      model.Meta{Title: "翻译字典", Icon: "reading"},
		},
	}
	utils.RegisterMenus(entities...)
}
