package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type ImageSearch struct {
	Type string `json:"type" form:"type"`
	Name string `json:"name" form:"name"`
	request.PageInfo
}
