package api

import (
	"net/url"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/translist/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Job = new(job)

type job struct{}

// UploadAndTranslate 上传Excel到对象存储并翻译
func (a *job) UploadAndTranslate(c *gin.Context) {
	dictName := c.PostForm("dictName")
	if dictName == "" {
		response.FailWithMessage("字典名称不能为空", c)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请上传Excel文件", c)
		return
	}

	job, err := serviceJob.UploadAndTranslate(c.Request.Context(), dictName, file)
	if err != nil {
		global.GVA_LOG.Error("翻译失败!", zap.Error(err))
		if job != nil {
			response.OkWithDetailed(job, "翻译失败: "+err.Error(), c)
			return
		}
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "翻译成功", c)
}

// DeleteJob 删除翻译任务
func (a *job) DeleteJob(c *gin.Context) {
	ID := c.Query("ID")
	if err := serviceJob.DeleteJob(c.Request.Context(), ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteJobByIds 批量删除
func (a *job) DeleteJobByIds(c *gin.Context) {
	IDs := c.QueryArray("IDs[]")
	if err := serviceJob.DeleteJobByIds(c.Request.Context(), IDs); err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// FindJob 查询单个任务
func (a *job) FindJob(c *gin.Context) {
	ID := c.Query("ID")
	job, err := serviceJob.GetJob(ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(job, c)
}

// GetJobList 分页列表
func (a *job) GetJobList(c *gin.Context) {
	var pageInfo request.JobSearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceJob.GetJobList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
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

// Retranslate 重新翻译
func (a *job) Retranslate(c *gin.Context) {
	ID := c.Query("ID")
	job, err := serviceJob.Retranslate(c.Request.Context(), ID)
	if err != nil {
		global.GVA_LOG.Error("重新翻译失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "重新翻译成功", c)
}

// ExportExcel 导出翻译结果
func (a *job) ExportExcel(c *gin.Context) {
	ID := c.Query("ID")
	path, fileName, cleanup, err := serviceJob.ExportFile(c.Request.Context(), ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer cleanup()
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(fileName))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.FileAttachment(path, fileName)
}
