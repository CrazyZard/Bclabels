package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type JobSearch struct {
	request.PageInfo
	DictName string `json:"dictName" form:"dictName"`
	Status   string `json:"status" form:"status"`
	FileName string `json:"fileName" form:"fileName"`
}
