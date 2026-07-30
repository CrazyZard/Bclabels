package initialize

import (
	"context"

	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
)

func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/image/createImage", Description: "新增图片", ApiGroup: "图片管理", Method: "POST"},
		{Path: "/image/deleteImage", Description: "删除图片", ApiGroup: "图片管理", Method: "DELETE"},
		{Path: "/image/deleteImageByIds", Description: "批量删除图片", ApiGroup: "图片管理", Method: "DELETE"},
		{Path: "/image/updateImage", Description: "更新图片", ApiGroup: "图片管理", Method: "PUT"},
		{Path: "/image/findImage", Description: "根据ID获取图片", ApiGroup: "图片管理", Method: "GET"},
		{Path: "/image/getImageList", Description: "分页获取图片列表", ApiGroup: "图片管理", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
