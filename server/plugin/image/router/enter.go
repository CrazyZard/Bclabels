package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/image/api"

var (
	Router    = new(router)
	apiImage = api.Image
)

type router struct{ Image image }
