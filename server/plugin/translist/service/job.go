package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	dictService "github.com/flipped-aurora/gin-vue-admin/server/plugin/dict/service"
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

func (s *job) DeleteJob(ctx context.Context, ID string) error {
	var j model.Job
	if err := global.GVA_DB.Where("id = ?", ID).First(&j).Error; err != nil {
		return err
	}
	deleteOssObject(ctx, j.SourceKey)
	deleteOssObject(ctx, j.ResultKey)
	return global.GVA_DB.Delete(&model.Job{}, "id = ?", ID).Error
}

func (s *job) DeleteJobByIds(ctx context.Context, IDs []string) error {
	var jobs []model.Job
	if err := global.GVA_DB.Where("id in ?", IDs).Find(&jobs).Error; err != nil {
		return err
	}
	for _, j := range jobs {
		deleteOssObject(ctx, j.SourceKey)
		deleteOssObject(ctx, j.ResultKey)
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

// UploadAndTranslate 上传 Excel 到对象存储并立即翻译
func (s *job) UploadAndTranslate(ctx context.Context, dictName string, fileHeader *multipart.FileHeader) (*model.Job, error) {
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

	safeName := sanitizeFileName(fileHeader.Filename)

	sourceURL, sourceKey, err := uploadMultipartToTranslist(ctx, fileHeader, "src")
	if err != nil {
		return nil, fmt.Errorf("上传源文件到对象存储失败: %w", err)
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		deleteOssObject(ctx, sourceKey)
		return nil, errors.New("读取上传文件失败")
	}
	srcBytes, readErr := io.ReadAll(srcFile)
	srcFile.Close()
	if readErr != nil {
		deleteOssObject(ctx, sourceKey)
		return nil, errors.New("读取上传文件失败")
	}

	job := &model.Job{
		FileName:   fileHeader.Filename,
		DictName:   dictName,
		Status:     "processing",
		SourcePath: sourceURL,
		SourceKey:  sourceKey,
	}
	if err := global.GVA_DB.Create(job).Error; err != nil {
		deleteOssObject(ctx, sourceKey)
		return nil, err
	}

	resultBytes, translated, missingWords, totalRows, err := s.translateExcelBytes(dictName, srcBytes)
	if err != nil {
		job.Status = "failed"
		job.ErrorMsg = err.Error()
		_ = global.GVA_DB.Model(job).Updates(map[string]interface{}{
			"status":    job.Status,
			"error_msg": job.ErrorMsg,
		}).Error
		return job, err
	}

	resultName := strings.TrimSuffix(safeName, filepath.Ext(safeName)) + "_translated" + ext
	resultURL, resultKey, err := uploadBytesToTranslist(ctx, resultBytes, resultName, "out")
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
	job.ResultPath = resultURL
	job.ResultKey = resultKey
	if err := global.GVA_DB.Model(job).Select(
		"status", "total_rows", "translated_cells", "missing_count", "missing_words", "result_path", "result_key",
	).Updates(map[string]interface{}{
		"status":           job.Status,
		"total_rows":       job.TotalRows,
		"translated_cells": job.TranslatedCells,
		"missing_count":    job.MissingCount,
		"missing_words":    job.MissingWords,
		"result_path":      job.ResultPath,
		"result_key":       job.ResultKey,
	}).Error; err != nil {
		return job, fmt.Errorf("保存翻译结果失败: %w", err)
	}

	return job, nil
}

// Retranslate 从对象存储拉取源文件重新翻译
func (s *job) Retranslate(ctx context.Context, ID string) (*model.Job, error) {
	j, err := s.GetJob(ID)
	if err != nil {
		return nil, err
	}
	if j.SourcePath == "" && j.SourceKey == "" {
		return nil, errors.New("源文件不存在")
	}

	srcPath, cleanup, err := fetchObjectToTemp(ctx, j.SourcePath, j.SourceKey)
	if err != nil {
		return nil, errors.New("源文件已丢失，请重新上传: " + err.Error())
	}
	defer cleanup()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, errors.New("读取源文件失败")
	}
	srcBytes, err := io.ReadAll(srcFile)
	srcFile.Close()
	if err != nil {
		return nil, errors.New("读取源文件失败")
	}

	_ = global.GVA_DB.Model(&j).Updates(map[string]interface{}{
		"status":    "processing",
		"error_msg": "",
	}).Error

	resultBytes, translated, missingWords, totalRows, err := s.translateExcelBytes(j.DictName, srcBytes)
	if err != nil {
		j.Status = "failed"
		j.ErrorMsg = err.Error()
		_ = global.GVA_DB.Model(&j).Updates(map[string]interface{}{
			"status":    j.Status,
			"error_msg": j.ErrorMsg,
		}).Error
		return &j, err
	}

	ext := strings.ToLower(filepath.Ext(j.FileName))
	if ext == "" {
		ext = ".xlsx"
	}
	resultName := strings.TrimSuffix(sanitizeFileName(j.FileName), filepath.Ext(sanitizeFileName(j.FileName))) + "_translated" + ext
	resultURL, resultKey, err := uploadBytesToTranslist(ctx, resultBytes, resultName, "out")
	if err != nil {
		j.Status = "failed"
		j.ErrorMsg = err.Error()
		_ = global.GVA_DB.Model(&j).Updates(map[string]interface{}{
			"status":    j.Status,
			"error_msg": j.ErrorMsg,
		}).Error
		return &j, err
	}

	// 删除旧结果对象（忽略失败）
	if j.ResultKey != "" && j.ResultKey != resultKey {
		deleteOssObject(ctx, j.ResultKey)
	}

	j.Status = "success"
	j.TotalRows = totalRows
	j.TranslatedCells = translated
	j.MissingCount = len(missingWords)
	j.MissingWords = mustJSON(missingWords)
	j.ResultPath = resultURL
	j.ResultKey = resultKey
	if err := global.GVA_DB.Model(&j).Select(
		"status", "total_rows", "translated_cells", "missing_count", "missing_words", "result_path", "result_key", "error_msg",
	).Updates(map[string]interface{}{
		"status":           j.Status,
		"total_rows":       j.TotalRows,
		"translated_cells": j.TranslatedCells,
		"missing_count":    j.MissingCount,
		"missing_words":    j.MissingWords,
		"result_path":      j.ResultPath,
		"result_key":       j.ResultKey,
		"error_msg":        "",
	}).Error; err != nil {
		return &j, fmt.Errorf("保存翻译结果失败: %w", err)
	}
	return &j, nil
}

func (s *job) translateExcelBytes(dictName string, srcBytes []byte) (result []byte, translated int, missingWords []string, totalRows int, err error) {
	dict, err := s.loadDict(dictName)
	if err != nil {
		return nil, 0, nil, 0, err
	}

	xlsx, err := excelize.OpenReader(bytes.NewReader(srcBytes))
	if err != nil {
		return nil, 0, nil, 0, errors.New("无法解析 Excel 文件: " + err.Error())
	}
	defer xlsx.Close()

	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return nil, 0, nil, 0, errors.New("Excel 文件中没有工作表")
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
		return nil, 0, nil, 0, errors.New("未识别到可翻译列，请确认表头含「xxx英文/俄文/阿语」等列")
	}

	buf, err := xlsx.WriteToBuffer()
	if err != nil {
		return nil, 0, nil, 0, fmt.Errorf("生成翻译结果失败: %w", err)
	}
	return buf.Bytes(), translated, dict.MissList(), totalRows, nil
}

func (s *job) loadDict(dictName string) (*dictService.Dict, error) {
	return dictService.LoadDict(dictName)
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

// ExportFile 返回可下载的结果文件本地路径；cleanup 在响应结束后调用。
func (s *job) ExportFile(ctx context.Context, ID string) (localPath string, fileName string, cleanup func(), err error) {
	noop := func() {}
	j, err := s.GetJob(ID)
	if err != nil {
		return "", "", noop, err
	}
	if j.Status != "success" {
		return "", "", noop, errors.New("任务尚未翻译成功，无法导出")
	}
	if j.ResultPath == "" && j.ResultKey == "" {
		return "", "", noop, errors.New("结果文件不存在")
	}
	localPath, cleanup, err = fetchObjectToTemp(ctx, j.ResultPath, j.ResultKey)
	if err != nil {
		return "", "", noop, errors.New("结果文件已丢失，请重新翻译: " + err.Error())
	}
	name := strings.TrimSuffix(j.FileName, filepath.Ext(j.FileName)) + "_已翻译" + filepath.Ext(j.FileName)
	return localPath, name, cleanup, nil
}
