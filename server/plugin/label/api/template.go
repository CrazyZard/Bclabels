package api

import (
	"encoding/json"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

var Template = new(template)

type template struct{}

type saveReq struct {
	Name               string `json:"name"`
	LabelWidth         int    `json:"labelWidth"`
	LabelHeight        int    `json:"labelHeight"`
	HeadSeam           int    `json:"headSeam"`
	MarginLR           int    `json:"marginLR"`
	NeedsTranslation   bool   `json:"needsTranslation"`
	DictName           string `json:"dictName"`
	TranslateLangs     string `json:"translateLangs"`
	Elements           string `json:"elements"`
	TranslatedElements string `json:"translatedElements"`
}

// SaveTemplate 保存标签模板
func (a *template) SaveTemplate(c *gin.Context) {
	var req saveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Name == "" {
		response.FailWithMessage("模板名称不能为空", c)
		return
	}
	if err := serviceTemplate.Save(req.Name, req.LabelWidth, req.LabelHeight, req.HeadSeam, req.MarginLR, req.NeedsTranslation, req.DictName, req.TranslateLangs, req.Elements, req.TranslatedElements); err != nil {
		global.GVA_LOG.Error("保存模板失败!", zap.Error(err))
		response.FailWithMessage("保存失败", c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// LoadTemplate 加载标签模板
func (a *template) LoadTemplate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.FailWithMessage("模板名称不能为空", c)
		return
	}
	t, err := serviceTemplate.Load(name)
	if err != nil {
		global.GVA_LOG.Error("加载模板失败!", zap.Error(err))
		response.FailWithMessage("模板不存在", c)
		return
	}
	response.OkWithData(t, c)
}

// ListTemplate 模板列表
func (a *template) ListTemplate(c *gin.Context) {
	list, err := serviceTemplate.List()
	if err != nil {
		global.GVA_LOG.Error("获取模板列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(list, c)
}

// DeleteTemplate 删除标签模板
func (a *template) DeleteTemplate(c *gin.Context) {
	name := c.Query("name")
	if err := serviceTemplate.Delete(name); err != nil {
		global.GVA_LOG.Error("删除模板失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// TranslateText 翻译文本（查询字典）
func (a *template) TranslateText(c *gin.Context) {
	dictName := c.Query("dictName")
	chinese := c.Query("chinese")
	lang := c.Query("lang")
	if dictName == "" || chinese == "" || lang == "" {
		response.FailWithMessage("参数不全", c)
		return
	}
	result, err := serviceTemplate.TranslateText(dictName, chinese, lang)
	if err != nil {
		response.OkWithDetailed(map[string]string{"translated": chinese, "notFound": "true"}, "字典中未找到该词条", c)
		return
	}
	response.OkWithDetailed(map[string]string{"translated": result}, "翻译成功", c)
}

// PublishTemplate 发布模板
func (a *template) PublishTemplate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.FailWithMessage("模板名称不能为空", c)
		return
	}
	if err := serviceTemplate.Publish(name); err != nil {
		global.GVA_LOG.Error("发布模板失败!", zap.Error(err))
		response.FailWithMessage("发布失败", c)
		return
	}
	response.OkWithMessage("发布成功", c)
}

// UnpublishTemplate 取消发布模板
func (a *template) UnpublishTemplate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.FailWithMessage("模板名称不能为空", c)
		return
	}
	if err := serviceTemplate.Unpublish(name); err != nil {
		global.GVA_LOG.Error("取消发布失败!", zap.Error(err))
		response.FailWithMessage("取消发布失败", c)
		return
	}
	response.OkWithMessage("已取消发布", c)
}

// ListPublishedTemplate 获取已发布模板列表
func (a *template) ListPublishedTemplate(c *gin.Context) {
	list, err := serviceTemplate.ListPublished()
	if err != nil {
		global.GVA_LOG.Error("获取已发布模板列表失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(list, c)
}

// DownloadBatchTemplate 下载批量导入Excel模板
func (a *template) DownloadBatchTemplate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		response.FailWithMessage("模板名称不能为空", c)
		return
	}
	tpl, err := serviceTemplate.Load(name)
	if err != nil {
		global.GVA_LOG.Error("加载模板失败!", zap.Error(err))
		response.FailWithMessage("模板不存在", c)
		return
	}

	// 解析正面元素JSON，提取所有text元素的key作为Excel表头
	var elements []map[string]interface{}
	var headers []string
	if err := json.Unmarshal([]byte(tpl.Elements), &elements); err == nil {
		seen := make(map[string]bool)
		for _, el := range elements {
			if elType, ok := el["type"].(string); ok && elType == "text" {
				if key, ok := el["key"].(string); ok && key != "" && !seen[key] {
					headers = append(headers, key)
					seen[key] = true
				}
			}
		}
	}

	if len(headers) == 0 {
		response.FailWithMessage("模板中没有可填充的文本字段", c)
		return
	}

	f := excelize.NewFile()
	sheet := "Sheet1"
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}
	// 添加示例行
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, fmt.Sprintf("示例%s", header))
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		global.GVA_LOG.Error("生成Excel失败!", zap.Error(err))
		response.FailWithMessage("生成Excel失败", c)
		return
	}

	filename := fmt.Sprintf("%s_批量导入模板.xlsx", name)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
