package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/service"

var (
	Api          = new(api)
	serviceEntry = service.Service.Entry
)

type api struct{ Entry entry }
