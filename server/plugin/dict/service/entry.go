package service

import (
	"errors"
	"io"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/model/request"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var Entry = new(entry)

type entry struct{}

func (s *entry) CreateEntry(entry *model.Entry) error {
	return global.GVA_DB.Create(entry).Error
}

func (s *entry) DeleteEntry(ID string) error {
	return global.GVA_DB.Delete(&model.Entry{}, "id = ?", ID).Error
}

func (s *entry) DeleteEntryByIds(IDs []string) error {
	return global.GVA_DB.Delete(&[]model.Entry{}, "id in ?", IDs).Error
}

func (s *entry) UpdateEntry(entry model.Entry) error {
	return global.GVA_DB.Model(&model.Entry{}).Where("id = ?", entry.ID).Updates(&entry).Error
}

func (s *entry) GetEntry(ID string) (model.Entry, error) {
	var entry model.Entry
	err := global.GVA_DB.Where("id = ?", ID).First(&entry).Error
	return entry, err
}

func (s *entry) GetEntryList(info request.EntrySearch) (list []model.Entry, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&model.Entry{})

	if info.DictName != "" {
		db = db.Where("dict_name = ?", info.DictName)
	}
	if info.Chinese != "" {
		db = db.Where("chinese LIKE ?", "%"+info.Chinese+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	err = db.Order("id DESC").Find(&list).Error
	return
}

type ImportResult struct {
	Success  int `json:"success"`
	Skip     int `json:"skip"`
	Fail     int `json:"fail"`
}

func (s *entry) ImportExcel(dictName string, file io.Reader) (*ImportResult, error) {
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		return nil, errors.New("无法解析 Excel 文件: " + err.Error())
	}
	defer xlsx.Close()

	result := &ImportResult{}
	sheets := xlsx.GetSheetList()

	if len(sheets) == 0 {
		return nil, errors.New("Excel 文件中没有工作表")
	}

	colMap := s.getColumnMap(dictName)

	for _, sheetName := range sheets {
		rows, err := xlsx.GetRows(sheetName)
		if err != nil {
			continue
		}
		if len(rows) < 2 {
			continue
		}

		// 跳过表头，从第2行开始
		for i := 1; i < len(rows); i++ {
			row := rows[i]
			if len(row) == 0 {
				continue
			}

			chinese := strings.TrimSpace(s.getCell(row, 0))
			if chinese == "" {
				continue
			}

			entry := model.Entry{
				DictName: dictName,
				Chinese:  chinese,
			}

			for _, cm := range colMap {
				val := strings.TrimSpace(s.getCell(row, cm.colIndex))
				switch cm.field {
				case "english":
					entry.English = val
				case "russian":
					entry.Russian = val
				case "arabic":
					entry.Arabic = val
				case "indonesian":
					entry.Indonesian = val
				}
			}

			// 查重：同一字典下中文唯一
			var existing model.Entry
			err := global.GVA_DB.Where("dict_name = ? AND chinese = ?", dictName, chinese).First(&existing).Error
			if err == nil {
				// 已存在，更新
				entry.ID = existing.ID
				entry.CreatedAt = existing.CreatedAt
				if uErr := global.GVA_DB.Model(&existing).Updates(map[string]interface{}{
					"english":    entry.English,
					"russian":    entry.Russian,
					"arabic":     entry.Arabic,
					"indonesian": entry.Indonesian,
				}).Error; uErr != nil {
					result.Fail++
				} else {
					result.Success++
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在，插入
				if cErr := global.GVA_DB.Create(&entry).Error; cErr != nil {
					result.Fail++
				} else {
					result.Success++
				}
			} else {
				result.Fail++
			}
		}
	}

	return result, nil
}

type colMapping struct {
	colIndex int
	field    string
}

func (s *entry) getColumnMap(dictName string) []colMapping {
	switch dictName {
	case "森马":
		// 森马: A-中文, B-英语, C-印尼语, D-俄语
		return []colMapping{
			{colIndex: 1, field: "english"},
			{colIndex: 2, field: "indonesian"},
			{colIndex: 3, field: "russian"},
		}
	default:
		// 巴拉 (及默认): A-中文, B-英文, C-俄文, D-阿语
		return []colMapping{
			{colIndex: 1, field: "english"},
			{colIndex: 2, field: "russian"},
			{colIndex: 3, field: "arabic"},
		}
	}
}

func (s *entry) getCell(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

// LoadDictionary 加载指定字典的全部条目（用于前端批量翻译）
func (s *entry) LoadDictionary(dictName string) ([]model.Entry, error) {
	var entries []model.Entry
	err := global.GVA_DB.Where("dict_name = ?", dictName).Find(&entries).Error
	return entries, err
}
