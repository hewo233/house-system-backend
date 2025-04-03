package OSS

import (
	"context"
	"fmt"
	"github.com/hewo233/house-system-backend/shared/consts"
	"github.com/minio/minio-go/v7"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"
)

func GetFileURL(objectName string) string {
	url := fmt.Sprintf("%s/%s/%s", endpoint, bucket, objectName)
	if useSSL {
		return "https://" + url
	}
	return "http://" + url
}

func UploadFileToOSS(ctx context.Context, category string, objectName string, fileReader io.Reader, fileSize int64, contentType string) (string, error) {
	// 构建完整的对象名（包含分类路径）
	fullObjectName := objectName
	if category != "" {
		fullObjectName = filepath.Join(category, objectName)
	}

	// 上传文件到 Minio
	_, err := minioClient.PutObject(ctx, bucket, fullObjectName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to OSS: %w", err)
	}

	return GetFileURL(fullObjectName), nil
}

func UploadImageToOSS(ctx context.Context, file *multipart.FileHeader) (string, error) {

	const maxFileSize = consts.TreeMB
	if file.Size > maxFileSize {
		return "", fmt.Errorf("文件大小超过限制，最大允许 10MB")
	}

	category := "images"

	// 根据文件扩展名设置 ContentType
	ext := filepath.Ext(file.Filename)
	contentType := "application/octet-stream"

	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".bmp":
		contentType = "image/bmp"
	case ".svg":
		contentType = "image/svg+xml"
	default:
		return "", fmt.Errorf("not suppot image type：%s，only support jpg/jpeg/png/gif/webp/bmp/svg", ext)
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("file to open image: %w", err)
	}
	defer src.Close()

	// 生成唯一的文件名
	objectName := fmt.Sprintf("%s-%s", time.Now().Format("20060102150405"), file.Filename)

	return UploadFileToOSS(ctx, category, objectName, src, file.Size, contentType)
}
