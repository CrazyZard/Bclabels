package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Template = new(template)

type template struct{}

func (r *template) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		group := private.Group("label").Use(middleware.OperationRecord())
		group.POST("saveTemplate", apiTemplate.SaveTemplate)
		group.DELETE("deleteTemplate", apiTemplate.DeleteTemplate)
		group.POST("publishTemplate", apiTemplate.PublishTemplate)
		group.POST("unpublishTemplate", apiTemplate.UnpublishTemplate)
	}
	{
		group := private.Group("label")
		group.GET("loadTemplate", apiTemplate.LoadTemplate)
		group.GET("listTemplate", apiTemplate.ListTemplate)
		group.GET("listPublishedTemplate", apiTemplate.ListPublishedTemplate)
		group.GET("translateText", apiTemplate.TranslateText)
		group.GET("downloadBatchTemplate", apiTemplate.DownloadBatchTemplate)
	}
}
