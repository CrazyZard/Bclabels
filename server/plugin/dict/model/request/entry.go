package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type EntrySearch struct {
	DictName string `json:"dictName" form:"dictName"`
	Chinese  string `json:"chinese" form:"chinese"`
	request.PageInfo
}
