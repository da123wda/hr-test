package service

import (
	"errors"
	"fmt"
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/cfg"
	"hr-dynamic-form-engine/src/dao"
	"hr-dynamic-form-engine/src/dto"
	"hr-dynamic-form-engine/src/model"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileServer struct{}

var allowExts = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

const maxUploadSize = 20 * 1024 * 1024 // 20MB

func (FileServer) UploadService(fileHeader *multipart.FileHeader, formId, uploadedBy string) *dto.UploadFileResp {
	if fileHeader.Size > maxUploadSize {
		cbase.PanicMsg("文件大小超出限制 (最大 20MB)")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowExts[ext] {
		cbase.PanicMsg(fmt.Sprintf("不支持的文件格式: %s", ext))
	}

	src, err := fileHeader.Open()
	if err != nil {
		cbase.PanicMsg("打开上传文件失败: " + err.Error())
	}
	defer src.Close()

	// 嗅探 MIME 类型
	buffer := make([]byte, 512)
	n, _ := src.Read(buffer)
	mimeType := http.DetectContentType(buffer[:n])
	_, _ = src.Seek(0, io.SeekStart)

	if strings.Contains(mimeType, "application/x-executable") ||
		strings.Contains(mimeType, "application/x-dosexec") ||
		strings.Contains(mimeType, "application/x-sh") {
		cbase.PanicMsg("安全拦截：禁止上传可执行文件")
	}

	_ = os.MkdirAll(cfg.Cfg.UploadDir, 0755)
	fileId := "file_" + uuid.New().String()
	storedName := fileId + ext
	dstPath := filepath.Join(cfg.Cfg.UploadDir, storedName)

	dst, err := os.Create(dstPath)
	if err != nil {
		cbase.PanicMsg("创建存储文件失败: " + err.Error())
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		cbase.PanicMsg("写入磁盘文件失败: " + err.Error())
	}

	sha256Hash, _ := calcFileSha256(src)
	downloadUrl := fmt.Sprintf("/api/v1/files/download/%s", fileId)

	meta := &model.FileMetadata{
		FileId:       fileId,
		OriginalName: fileHeader.Filename,
		StoredName:   storedName,
		FileSize:     written,
		MimeType:     mimeType,
		Sha256:       sha256Hash,
		DownloadUrl:  downloadUrl,
		FormId:       formId,
		UploadedBy:   uploadedBy,
		CreatedAt:    time.Now(),
	}

	if err := dao.FileDbDefault().InsertOne(meta); err != nil {
		_ = os.Remove(dstPath)
		cbase.PanicMsg("保存文件元数据失败: " + err.Error())
	}

	return &dto.UploadFileResp{
		FileId:       meta.FileId,
		OriginalName: meta.OriginalName,
		FileSize:     meta.FileSize,
		MimeType:     meta.MimeType,
		DownloadUrl:  meta.DownloadUrl,
	}
}

func (FileServer) GetFileForDownload(fileId string) (string, *model.FileMetadata, error) {
	meta, err := dao.FileDbDefault().GetById(fileId)
	if err != nil {
		return "", nil, err
	}
	filePath := filepath.Join(cfg.Cfg.UploadDir, meta.StoredName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil, errors.New("物理文件不存在")
	}
	return filePath, meta, nil
}
