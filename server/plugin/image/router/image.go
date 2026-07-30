package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Image = new(image)

type image struct{}

func (r *image) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		group := private.Group("image").Use(middleware.OperationRecord())
		group.POST("createImage", apiImage.CreateImage)
		group.DELETE("deleteImage", apiImage.DeleteImage)
		group.DELETE("deleteImageByIds", apiImage.DeleteImageByIds)
		group.PUT("updateImage", apiImage.UpdateImage)
	}
	{
		group := private.Group("image")
		group.GET("findImage", apiImage.FindImage)
		group.GET("getImageList", apiImage.GetImageList)
	}
}
