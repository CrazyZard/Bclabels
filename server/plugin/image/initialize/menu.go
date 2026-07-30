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
			Path:      "imageManager",
			Name:      "imageManager",
			Hidden:    false,
			Component: "plugin/image/view/index.vue",
			Sort:      8,
			Meta:      model.Meta{Title: "图片管理", Icon: "picture"},
		},
	}
	utils.RegisterMenus(entities...)
}
