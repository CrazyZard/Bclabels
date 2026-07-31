package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/upload"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// 翻译列表文件在对象存储中的统一目录
const translistOssDir = "translist"

func buildTranslistKey(fileName, kind string) string {
	safe := sanitizeFileName(fileName)
	ts := time.Now().Format("20060102150405")
	return fmt.Sprintf("%s/%s_%s_%s", translistOssDir, ts, kind, safe)
}

func publicURLForKey(key string) string {
	key = strings.TrimPrefix(key, "/")
	switch global.GVA_CONFIG.System.OssType {
	case "local", "":
		return strings.TrimRight(global.GVA_CONFIG.Local.Path, "/") + "/" + key
	case "qiniu":
		img := strings.TrimRight(strings.TrimSpace(global.GVA_CONFIG.Qiniu.ImgPath), "/")
		if img == "" {
			return key
		}
		return img + "/" + key
	default:
		img := strings.TrimRight(strings.TrimSpace(global.GVA_CONFIG.Qiniu.ImgPath), "/")
		if img != "" {
			return img + "/" + key
		}
		return key
	}
}

// uploadReaderToTranslist 上传到对象存储 translist/ 目录，返回可访问 URL 与最终 key。
func uploadReaderToTranslist(ctx context.Context, r io.Reader, size int64, key string) (fileURL, finalKey string, err error) {
	key = strings.TrimPrefix(filepath.ToSlash(key), "/")
	if !strings.HasPrefix(key, translistOssDir+"/") {
		key = translistOssDir + "/" + key
	}
	finalKey = key

	switch global.GVA_CONFIG.System.OssType {
	case "local", "":
		storePath := filepath.Join(global.GVA_CONFIG.Local.StorePath, filepath.FromSlash(key))
		if err = os.MkdirAll(filepath.Dir(storePath), os.ModePerm); err != nil {
			return "", "", fmt.Errorf("创建本地目录失败: %w", err)
		}
		out, createErr := os.Create(storePath)
		if createErr != nil {
			return "", "", fmt.Errorf("创建本地文件失败: %w", createErr)
		}
		if _, err = io.Copy(out, r); err != nil {
			out.Close()
			_ = os.Remove(storePath)
			return "", "", fmt.Errorf("写入本地文件失败: %w", err)
		}
		out.Close()
		return publicURLForKey(key), finalKey, nil

	case "qiniu":
		if !qiniuConfigured() {
			return "", "", fmt.Errorf("七牛配置不完整")
		}
		putPolicy := storage.PutPolicy{Scope: global.GVA_CONFIG.Qiniu.Bucket}
		mac := qbox.NewMac(global.GVA_CONFIG.Qiniu.AccessKey, global.GVA_CONFIG.Qiniu.SecretKey)
		upToken := putPolicy.UploadToken(mac)
		formUploader := storage.NewFormUploader(qiniuStorageConfig())
		ret := storage.PutRet{}
		putExtra := storage.PutExtra{}
		if err = formUploader.Put(ctx, &ret, upToken, key, r, size, &putExtra); err != nil {
			return "", "", fmt.Errorf("七牛上传失败: %w", err)
		}
		finalKey = ret.Key
		return publicURLForKey(finalKey), finalKey, nil

	default:
		tmp, tmpErr := os.CreateTemp("", "translist-*"+filepath.Ext(key))
		if tmpErr != nil {
			return "", "", tmpErr
		}
		tmpPath := tmp.Name()
		defer func() { _ = os.Remove(tmpPath) }()
		if _, err = io.Copy(tmp, r); err != nil {
			tmp.Close()
			return "", "", err
		}
		tmp.Close()
		header, cleanup, hErr := upload.BuildFileHeader(tmpPath, "file", filepath.Base(key))
		if hErr != nil {
			return "", "", hErr
		}
		defer cleanup()
		fileURL, uploadedKey, err := upload.NewOss().UploadFile(ctx, header)
		if err != nil {
			return "", "", err
		}
		return fileURL, uploadedKey, nil
	}
}

func uploadMultipartToTranslist(ctx context.Context, fileHeader *multipart.FileHeader, kind string) (fileURL, key string, err error) {
	key = buildTranslistKey(fileHeader.Filename, kind)
	f, err := fileHeader.Open()
	if err != nil {
		return "", "", errors.New("读取上传文件失败")
	}
	defer f.Close()
	return uploadReaderToTranslist(ctx, f, fileHeader.Size, key)
}

func uploadBytesToTranslist(ctx context.Context, data []byte, fileName, kind string) (fileURL, key string, err error) {
	key = buildTranslistKey(fileName, kind)
	return uploadReaderToTranslist(ctx, bytes.NewReader(data), int64(len(data)), key)
}

func deleteOssObject(ctx context.Context, key string) {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if key == "" {
		return
	}
	_ = upload.NewOss().DeleteFile(ctx, key)
}

func resolveObjectKey(key, fileURL string) string {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if key != "" {
		return key
	}
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return ""
	}
	if strings.HasPrefix(fileURL, "http://") || strings.HasPrefix(fileURL, "https://") {
		if u, err := url.Parse(fileURL); err == nil {
			return strings.TrimPrefix(u.Path, "/")
		}
	}
	return strings.TrimPrefix(filepath.ToSlash(fileURL), "/")
}

func qiniuConfigured() bool {
	return global.GVA_CONFIG.Qiniu.AccessKey != "" &&
		global.GVA_CONFIG.Qiniu.SecretKey != "" &&
		global.GVA_CONFIG.Qiniu.Bucket != ""
}

func qiniuStorageConfig() *storage.Config {
	cfg := &storage.Config{
		UseHTTPS:      global.GVA_CONFIG.Qiniu.UseHTTPS,
		UseCdnDomains: global.GVA_CONFIG.Qiniu.UseCdnDomains,
	}
	switch strings.ToLower(strings.ReplaceAll(global.GVA_CONFIG.Qiniu.Zone, "-", "")) {
	case "zonehuadong", "z0":
		cfg.Zone = &storage.ZoneHuadong
	case "zonehuabei", "z1":
		cfg.Zone = &storage.ZoneHuabei
	case "zonehuanan", "z2":
		cfg.Zone = &storage.ZoneHuanan
	case "zonebeimei", "na0":
		cfg.Zone = &storage.ZoneBeimei
	case "zonexinjiapo", "as0":
		cfg.Zone = &storage.ZoneXinjiapo
	}
	return cfg
}

// fetchObjectToTemp 从对象存储拉取到临时文件。cleanup 用完后必须调用。
func fetchObjectToTemp(ctx context.Context, fileURL, key string) (localPath string, cleanup func(), err error) {
	noop := func() {}
	objectKey := resolveObjectKey(key, fileURL)
	if objectKey == "" {
		return "", noop, fmt.Errorf("缺少对象存储 Key")
	}

	if global.GVA_CONFIG.System.OssType == "local" || global.GVA_CONFIG.System.OssType == "" {
		p := filepath.Join(global.GVA_CONFIG.Local.StorePath, filepath.FromSlash(objectKey))
		if _, statErr := os.Stat(p); statErr == nil {
			return p, noop, nil
		}
	}

	ext := filepath.Ext(objectKey)
	if ext == "" {
		ext = ".xlsx"
	}

	var errs []string

	if qiniuConfigured() && (global.GVA_CONFIG.System.OssType == "qiniu" || global.GVA_CONFIG.System.OssType == "") {
		path, clean, qErr := downloadQiniuToTemp(ctx, objectKey, ext)
		if qErr == nil {
			return path, clean, nil
		}
		errs = append(errs, qErr.Error())
	}

	downloadURL := ""
	if strings.HasPrefix(fileURL, "http://") || strings.HasPrefix(fileURL, "https://") {
		downloadURL = fileURL
	} else if img := strings.TrimRight(strings.TrimSpace(global.GVA_CONFIG.Qiniu.ImgPath), "/"); img != "" {
		downloadURL = img + "/" + objectKey
	}
	if downloadURL != "" {
		req, httpErr := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
		if httpErr != nil {
			errs = append(errs, httpErr.Error())
		} else {
			client := &http.Client{Timeout: 2 * time.Minute}
			resp, httpErr := client.Do(req)
			if httpErr != nil {
				errs = append(errs, httpErr.Error())
			} else {
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errs = append(errs, fmt.Sprintf("HTTP %d", resp.StatusCode))
				} else {
					return writeTempFromReader(resp.Body, ext)
				}
			}
		}
	}

	if len(errs) > 0 {
		return "", noop, fmt.Errorf("下载失败(key=%s): %s", objectKey, strings.Join(errs, "; "))
	}
	return "", noop, fmt.Errorf("无法下载文件(key=%s)，请检查对象存储配置", objectKey)
}

func downloadQiniuToTemp(ctx context.Context, key, ext string) (string, func(), error) {
	mac := qbox.NewMac(global.GVA_CONFIG.Qiniu.AccessKey, global.GVA_CONFIG.Qiniu.SecretKey)
	manager := storage.NewBucketManager(mac, qiniuStorageConfig())

	var domains []string
	if p := strings.TrimSpace(global.GVA_CONFIG.Qiniu.ImgPath); p != "" {
		domains = append(domains, strings.TrimRight(p, "/"))
	}

	out, err := manager.Get(global.GVA_CONFIG.Qiniu.Bucket, key, &storage.GetObjectInput{
		Context:         ctx,
		DownloadDomains: domains,
		PresignUrl:      true,
	})
	if err != nil {
		if out != nil {
			_ = out.Close()
		}
		return "", func() {}, fmt.Errorf("七牛下载失败: %w", err)
	}
	defer out.Close()
	return writeTempFromReader(out, ext)
}

func writeTempFromReader(r io.Reader, ext string) (string, func(), error) {
	noop := func() {}
	tmp, err := os.CreateTemp("", "translist-dl-*"+ext)
	if err != nil {
		return "", noop, err
	}
	tmpPath := tmp.Name()
	if _, err = io.Copy(tmp, r); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return "", noop, fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", noop, err
	}
	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}
