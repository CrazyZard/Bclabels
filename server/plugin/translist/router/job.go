package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Job = new(job)

type job struct{}

func (r *job) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		group := private.Group("translist").Use(middleware.OperationRecord())
		group.POST("uploadAndTranslate", apiJob.UploadAndTranslate)
		group.DELETE("deleteJob", apiJob.DeleteJob)
		group.DELETE("deleteJobByIds", apiJob.DeleteJobByIds)
		group.POST("retranslate", apiJob.Retranslate)
	}
	{
		group := private.Group("translist")
		group.GET("findJob", apiJob.FindJob)
		group.GET("getJobList", apiJob.GetJobList)
		group.GET("exportExcel", apiJob.ExportExcel)
	}
	_ = public
}
