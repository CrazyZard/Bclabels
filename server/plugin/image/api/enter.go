package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/image/service"

var (
	Api             = new(api)
	serviceImage = service.Image
)

type api struct{ Image image }
