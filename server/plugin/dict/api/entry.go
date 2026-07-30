package api

import (
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/model/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Entry = new(entry)

type entry struct{}

// CreateEntry 新增翻译条目
// @Tags Entry
// @Summary 新增翻译条目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Entry true "新增翻译条目"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /entry/createEntry [post]
func (a *entry) CreateEntry(c *gin.Context) {
	var entry model.Entry
	if err := c.ShouldBindJSON(&entry); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceEntry.CreateEntry(&entry); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		if strings.Contains(err.Error(), "Duplicate entry") {
			response.FailWithMessage("字典已存在", c)
			return
		}
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteEntry 删除翻译条目
// @Tags Entry
// @Summary 删除翻译条目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /entry/deleteEntry [delete]
func (a *entry) DeleteEntry(c *gin.Context) {
	ID := c.Query("ID")
	if err := serviceEntry.DeleteEntry(ID); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteEntryByIds 批量删除翻译条目
// @Tags Entry
// @Summary 批量删除翻译条目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /entry/deleteEntryByIds [delete]
func (a *entry) DeleteEntryByIds(c *gin.Context) {
	IDs := c.QueryArray("IDs[]")
	if err := serviceEntry.DeleteEntryByIds(IDs); err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateEntry 更新翻译条目
// @Tags Entry
// @Summary 更新翻译条目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body model.Entry true "更新翻译条目"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /entry/updateEntry [put]
func (a *entry) UpdateEntry(c *gin.Context) {
	var entry model.Entry
	if err := c.ShouldBindJSON(&entry); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := serviceEntry.UpdateEntry(entry); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		if strings.Contains(err.Error(), "Duplicate entry") {
			response.FailWithMessage("字典已存在", c)
			return
		}
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindEntry 根据ID查询翻译条目
// @Tags Entry
// @Summary 根据ID查询翻译条目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query model.Entry true "用id查询翻译条目"
// @Success 200 {object} response.Response{data=model.Entry,msg=string} "查询成功"
// @Router /entry/findEntry [get]
func (a *entry) FindEntry(c *gin.Context) {
	ID := c.Query("ID")
	reentry, err := serviceEntry.GetEntry(ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(reentry, c)
}

// GetEntryList 分页获取翻译条目列表
// @Tags Entry
// @Summary 分页获取翻译条目列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query request.EntrySearch true "分页获取翻译条目列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /entry/getEntryList [get]
func (a *entry) GetEntryList(c *gin.Context) {
	var pageInfo request.EntrySearch
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := serviceEntry.GetEntryList(pageInfo)
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

// ImportExcel 上传Excel导入翻译数据
// @Tags Entry
// @Summary 上传Excel导入翻译数据
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Param dictName formData string true "字典名称(巴拉/森马)"
// @Param file formData file true "Excel文件"
// @Success 200 {object} response.Response{data=service.ImportResult,msg=string} "导入成功"
// @Router /entry/importExcel [post]
func (a *entry) ImportExcel(c *gin.Context) {
	dictName := c.PostForm("dictName")
	if dictName == "" {
		response.FailWithMessage("字典名称不能为空", c)
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.FailWithMessage("请上传Excel文件", c)
		return
	}
	defer file.Close()

	result, err := serviceEntry.ImportExcel(dictName, file)
	if err != nil {
		global.GVA_LOG.Error("导入失败!", zap.Error(err))
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "导入完成", c)
}

// LoadDictionary 加载完整字典（前端批量翻译用）
func (a *entry) LoadDictionary(c *gin.Context) {
	dictName := c.Query("dictName")
	if dictName == "" {
		response.FailWithMessage("字典名称不能为空", c)
		return
	}
	entries, err := serviceEntry.LoadDictionary(dictName)
	if err != nil {
		global.GVA_LOG.Error("加载字典失败!", zap.Error(err))
		response.FailWithMessage("加载失败", c)
		return
	}
	response.OkWithData(entries, c)
}
