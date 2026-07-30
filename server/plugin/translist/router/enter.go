package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/translist/api"

var (
	Router = new(router)
	apiJob = api.Api.Job
)

type router struct{ Job job }
