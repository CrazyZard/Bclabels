package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	dictModel "github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/translist/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/translist/model/request"
	"github.com/xuri/excelize/v2"
)

var Job = new(job)

type job struct{}

type langSuffix struct {
	Suffix string
	Lang   string
}

var langSuffixes = []langSuffix{
	{Suffix: "印尼语", Lang: "indonesian"},
	{Suffix: "印尼文", Lang: "indonesian"},
	{Suffix: "英文", Lang: "english"},
	{Suffix: "俄文", Lang: "russian"},
	{Suffix: "阿语", Lang: "arabic"},
	{Suffix: "阿文", Lang: "arabic"},
}

type colPair struct {
	TargetIdx  int
	SourceIdx  int
	SourceName string
	Lang       string
}

func (s *job) CreateJob(job *model.Job) error {
	return global.GVA_DB.Create(job).Error
}

func (s *job) DeleteJob(ID string) error {
	var j model.Job
	if err := global.GVA_DB.Where("id = ?", ID).First(&j).Error; err != nil {
		return err
	}
	_ = removeFile(j.SourcePath)
	_ = removeFile(j.ResultPath)
	return global.GVA_DB.Delete(&model.Job{}, "id = ?", ID).Error
}

func (s *job) DeleteJobByIds(IDs []string) error {
	var jobs []model.Job
	if err := global.GVA_DB.Where("id in ?", IDs).Find(&jobs).Error; err != nil {
		return err
	}
	for _, j := range jobs {
		_ = removeFile(j.SourcePath)
		_ = removeFile(j.ResultPath)
	}
	return global.GVA_DB.Delete(&[]model.Job{}, "id in ?", IDs).Error
}

func (s *job) GetJob(ID string) (model.Job, error) {
	var j model.Job
	err := global.GVA_DB.Where("id = ?", ID).First(&j).Error
	return j, err
}

func (s *job) GetJobList(info request.JobSearch) (list []model.Job, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&model.Job{})

	if info.DictName != "" {
		db = db.Where("dict_name = ?", info.DictName)
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.FileName != "" {
		db = db.Where("file_name LIKE ?", "%"+info.FileName+"%")
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

// UploadAndTranslate 上传 Excel 并立即翻译
func (s *job) UploadAndTranslate(dictName string, fileHeader *multipart.FileHeader) (*model.Job, error) {
	if dictName == "" {
		return nil, errors.New("字典名称不能为空")
	}
	if fileHeader == nil {
		return nil, errors.New("请上传Excel文件")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".xlsx" && ext != ".xls" {
		return nil, errors.New("仅支持 .xlsx / .xls 文件")
	}

	storeDir := filepath.Join("uploads", "translist")
	if err := os.MkdirAll(storeDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	ts := time.Now().Format("20060102150405")
	safeName := sanitizeFileName(fileHeader.Filename)
	sourceRel := filepath.ToSlash(filepath.Join(storeDir, fmt.Sprintf("%s_src_%s", ts, safeName)))
	resultRel := filepath.ToSlash(filepath.Join(storeDir, fmt.Sprintf("%s_out_%s", ts, safeName)))

	srcFile, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("读取上传文件失败")
	}
	defer srcFile.Close()

	out, err := os.Create(sourceRel)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	if _, err = io.Copy(out, srcFile); err != nil {
		out.Close()
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	out.Close()

	job := &model.Job{
		FileName:   fileHeader.Filename,
		DictName:   dictName,
		Status:     "processing",
		SourcePath: sourceRel,
		ResultPath: resultRel,
	}
	if err := global.GVA_DB.Create(job).Error; err != nil {
		_ = os.Remove(sourceRel)
		return nil, err
	}

	translated, missingWords, totalRows, err := s.translateExcelFile(dictName, sourceRel, resultRel)
	if err != nil {
		job.Status = "failed"
		job.ErrorMsg = err.Error()
		_ = global.GVA_DB.Model(job).Updates(map[string]interface{}{
			"status":    job.Status,
			"error_msg": job.ErrorMsg,
		}).Error
		return job, err
	}

	job.Status = "success"
	job.TotalRows = totalRows
	job.TranslatedCells = translated
	job.MissingCount = len(missingWords)
	job.MissingWords = mustJSON(missingWords)
	if err := global.GVA_DB.Model(job).Select(
		"status", "total_rows", "translated_cells", "missing_count", "missing_words", "result_path",
	).Updates(map[string]interface{}{
		"status":           job.Status,
		"total_rows":       job.TotalRows,
		"translated_cells": job.TranslatedCells,
		"missing_count":    job.MissingCount,
		"missing_words":    job.MissingWords,
		"result_path":      job.ResultPath,
	}).Error; err != nil {
		return job, fmt.Errorf("保存翻译结果失败: %w", err)
	}

	return job, nil
}

// Retranslate 使用原文件重新翻译
func (s *job) Retranslate(ID string) (*model.Job, error) {
	j, err := s.GetJob(ID)
	if err != nil {
		return nil, err
	}
	if j.SourcePath == "" {
		return nil, errors.New("源文件不存在")
	}
	if _, err := os.Stat(j.SourcePath); err != nil {
		return nil, errors.New("源文件已丢失，请重新上传")
	}

	if j.ResultPath == "" {
		ts := time.Now().Format("20060102150405")
		j.ResultPath = filepath.ToSlash(filepath.Join("uploads", "translist", fmt.Sprintf("%s_out_%s", ts, sanitizeFileName(j.FileName))))
	}

	_ = global.GVA_DB.Model(&j).Updates(map[string]interface{}{
		"status":    "processing",
		"error_msg": "",
	}).Error

	translated, missingWords, totalRows, err := s.translateExcelFile(j.DictName, j.SourcePath, j.ResultPath)
	if err != nil {
		j.Status = "failed"
		j.ErrorMsg = err.Error()
		_ = global.GVA_DB.Model(&j).Updates(map[string]interface{}{
			"status":    j.Status,
			"error_msg": j.ErrorMsg,
		}).Error
		return &j, err
	}

	j.Status = "success"
	j.TotalRows = totalRows
	j.TranslatedCells = translated
	j.MissingCount = len(missingWords)
	j.MissingWords = mustJSON(missingWords)
	if err := global.GVA_DB.Model(&j).Select(
		"status", "total_rows", "translated_cells", "missing_count", "missing_words", "result_path", "error_msg",
	).Updates(map[string]interface{}{
		"status":           j.Status,
		"total_rows":       j.TotalRows,
		"translated_cells": j.TranslatedCells,
		"missing_count":    j.MissingCount,
		"missing_words":    j.MissingWords,
		"result_path":      j.ResultPath,
		"error_msg":        "",
	}).Error; err != nil {
		return &j, fmt.Errorf("保存翻译结果失败: %w", err)
	}
	return &j, nil
}

func (s *job) translateExcelFile(dictName, sourcePath, resultPath string) (translated int, missingWords []string, totalRows int, err error) {
	dict, err := s.loadDict(dictName)
	if err != nil {
		return 0, nil, 0, err
	}

	xlsx, err := excelize.OpenFile(sourcePath)
	if err != nil {
		return 0, nil, 0, errors.New("无法解析 Excel 文件: " + err.Error())
	}
	defer xlsx.Close()

	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return 0, nil, 0, errors.New("Excel 文件中没有工作表")
	}

	for _, sheetName := range sheets {
		rows, err := xlsx.GetRows(sheetName)
		if err != nil || len(rows) < 2 {
			continue
		}

		headers := rows[0]
		pairs := detectColPairs(headers)
		if len(pairs) == 0 {
			continue
		}

		for r := 1; r < len(rows); r++ {
			row := rows[r]
			if isEmptyRow(row) {
				continue
			}
			totalRows++
			for _, p := range pairs {
				src := getCell(row, p.SourceIdx)
				if strings.TrimSpace(src) == "" {
					continue
				}
				dst := dict.TranslateText(src, p.Lang)
				cell, _ := excelize.CoordinatesToCellName(p.TargetIdx+1, r+1)
				_ = xlsx.SetCellValue(sheetName, cell, dst)
				translated++
			}
		}
	}

	if translated == 0 && totalRows == 0 {
		return 0, nil, 0, errors.New("未识别到可翻译列，请确认表头含「xxx英文/俄文/阿语」等列")
	}

	if err := os.MkdirAll(filepath.Dir(resultPath), os.ModePerm); err != nil {
		return 0, nil, 0, err
	}
	if err := xlsx.SaveAs(resultPath); err != nil {
		return 0, nil, 0, fmt.Errorf("保存翻译结果失败: %w", err)
	}
	return translated, dict.MissList(), totalRows, nil
}

func (s *job) loadDict(dictName string) (*Dict, error) {
	var entries []dictModel.Entry
	if err := global.GVA_DB.Where("dict_name = ?", dictName).Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("字典「%s」为空，请先在翻译字典中导入词条", dictName)
	}
	d := &Dict{Entries: make(map[string]map[string]string, len(entries))}
	for _, e := range entries {
		m := map[string]string{}
		if e.English != "" {
			m["english"] = e.English
		}
		if e.Russian != "" {
			m["russian"] = e.Russian
		}
		if e.Arabic != "" {
			m["arabic"] = e.Arabic
		}
		if e.Indonesian != "" {
			m["indonesian"] = e.Indonesian
		}
		d.Entries[e.Chinese] = m
	}
	return d, nil
}

func detectColPairs(headers []string) []colPair {
	nameToIdx := map[string]int{}
	for i, h := range headers {
		nameToIdx[strings.TrimSpace(h)] = i
	}

	var pairs []colPair
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		for _, suf := range langSuffixes {
			if strings.HasSuffix(h, suf.Suffix) && h != suf.Suffix {
				srcName := strings.TrimSuffix(h, suf.Suffix)
				srcIdx, ok := nameToIdx[srcName]
				if !ok {
					continue
				}
				pairs = append(pairs, colPair{
					TargetIdx:  i,
					SourceIdx:  srcIdx,
					SourceName: srcName,
					Lang:       suf.Lang,
				})
				break
			}
		}
	}
	return pairs
}

func getCell(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

func isEmptyRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	var b strings.Builder
	for _, r := range name {
		switch {
		case r == '.' || r == '-' || r == '_' || r == '(' || r == ')':
			b.WriteRune(r)
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		case utf8.ValidRune(r) && r > 127:
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "file.xlsx"
	}
	return out
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil || v == nil {
		return "[]"
	}
	if b == nil || string(b) == "null" {
		return "[]"
	}
	return string(b)
}

func removeFile(path string) error {
	if path == "" {
		return nil
	}
	return os.Remove(path)
}

// ExportPath 返回可下载的结果文件路径
func (s *job) ExportPath(ID string) (absPath string, fileName string, err error) {
	j, err := s.GetJob(ID)
	if err != nil {
		return "", "", err
	}
	if j.Status != "success" {
		return "", "", errors.New("任务尚未翻译成功，无法导出")
	}
	if j.ResultPath == "" {
		return "", "", errors.New("结果文件不存在")
	}
	if _, err := os.Stat(j.ResultPath); err != nil {
		return "", "", errors.New("结果文件已丢失，请重新翻译")
	}
	name := strings.TrimSuffix(j.FileName, filepath.Ext(j.FileName)) + "_已翻译" + filepath.Ext(j.FileName)
	return j.ResultPath, name, nil
}
