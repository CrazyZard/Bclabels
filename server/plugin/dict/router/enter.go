package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/api"

var (
	Router   = new(router)
	apiEntry = api.Api.Entry
)

type router struct{ Entry entry }
