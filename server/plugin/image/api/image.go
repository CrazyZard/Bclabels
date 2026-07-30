package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/image/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/image/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Image = new(image)

type image struct{}

func (a *image) CreateImage(c *gin.Context) {
	var img model.Image
	if err := c.ShouldBindJSON(&img); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceImage.Create(&img); err != nil {
		global.GVA_LOG.Error("创建图片失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *image) DeleteImage(c *gin.Context) {
	id := c.Query("ID")
	if err := serviceImage.Delete(id); err != nil {
		global.GVA_LOG.Error("删除图片失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *image) DeleteImageByIds(c *gin.Context) {
	ids := c.QueryArray("IDs[]")
	if err := serviceImage.DeleteByIds(ids); err != nil {
		global.GVA_LOG.Error("批量删除图片失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

func (a *image) UpdateImage(c *gin.Context) {
	var img model.Image
	if err := c.ShouldBindJSON(&img); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceImage.Update(img); err != nil {
		global.GVA_LOG.Error("更新图片失败!", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *image) FindImage(c *gin.Context) {
	id := c.Query("ID")
	img, err := serviceImage.GetByID(id)
	if err != nil {
		global.GVA_LOG.Error("查询图片失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(img, c)
}

func (a *image) GetImageList(c *gin.Context) {
	var pageInfo request.ImageSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceImage.GetList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取图片列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
