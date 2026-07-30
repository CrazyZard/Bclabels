package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/label/api"

var (
	Router       = new(router)
	apiTemplate  = api.Api.Template
)

type router struct{ Template template }
