package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// Job 翻译任务
type Job struct {
	global.GVA_MODEL
	FileName        string `json:"fileName" gorm:"column:file_name;comment:原始文件名;size:255"`
	DictName        string `json:"dictName" gorm:"column:dict_name;comment:翻译字典;size:50;index"`
	Status          string `json:"status" gorm:"column:status;comment:状态(pending/processing/success/failed);size:20;index"`
	TotalRows       int    `json:"totalRows" gorm:"column:total_rows;comment:数据行数"`
	TranslatedCells int    `json:"translatedCells" gorm:"column:translated_cells;comment:成功翻译单元格数"`
	MissingCount    int    `json:"missingCount" gorm:"column:missing_count;comment:词典未命中词条数"`
	MissingWords    string `json:"missingWords" gorm:"column:missing_words;comment:未命中词条JSON;type:longtext"`
	SourcePath      string `json:"sourcePath" gorm:"column:source_path;comment:源文件URL或路径;size:500"`
	SourceKey       string `json:"sourceKey" gorm:"column:source_key;comment:源文件OSS Key;size:255"`
	ResultPath      string `json:"resultPath" gorm:"column:result_path;comment:结果文件URL或路径;size:500"`
	ResultKey       string `json:"resultKey" gorm:"column:result_key;comment:结果文件OSS Key;size:255"`
	ErrorMsg        string `json:"errorMsg" gorm:"column:error_msg;comment:错误信息;type:text"`
	Remark          string `json:"remark" gorm:"column:remark;comment:备注;size:255"`
}

func (Job) TableName() string {
	return "gva_translist_jobs"
}
