package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/translist/service"

var (
	Api        = new(api)
	serviceJob = service.Service.Job
)

type api struct{ Job job }
