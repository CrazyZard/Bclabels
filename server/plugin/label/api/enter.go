package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/label/service"

var (
	Api              = new(api)
	serviceTemplate  = service.Service.Template
)

type api struct{ Template template }
