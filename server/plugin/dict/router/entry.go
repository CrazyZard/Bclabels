package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Entry = new(entry)

type entry struct{}

func (r *entry) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		group := private.Group("entry").Use(middleware.OperationRecord())
		group.POST("createEntry", apiEntry.CreateEntry)
		group.DELETE("deleteEntry", apiEntry.DeleteEntry)
		group.DELETE("deleteEntryByIds", apiEntry.DeleteEntryByIds)
		group.PUT("updateEntry", apiEntry.UpdateEntry)
		group.POST("importExcel", apiEntry.ImportExcel)
	}
	{
		group := private.Group("entry")
		group.GET("findEntry", apiEntry.FindEntry)
		group.GET("getEntryList", apiEntry.GetEntryList)
		group.GET("loadDictionary", apiEntry.LoadDictionary)
	}
}
