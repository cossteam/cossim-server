package model

import (
	"github.com/minio/minio-go/v7"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func GetContentTypeOption(fileExt string) minio.PutObjectOptions {
	contentType := ""

	// 根据文件扩展名设置ContentType
	switch fileExt {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".wav":
		contentType = "audio/wav"
	case ".mp4":
		contentType = "video/mp4"
	case ".avi":
		contentType = "video/x-msvideo"
	// 添加其他文件类型的判断逻辑
	default:
		contentType = "application/octet-stream" // 默认使用二进制流的ContentType
	}

	options := minio.PutObjectOptions{
		ContentType: contentType,
	}
	return options
}

type FileType int

const (
	FileType_Voice FileType = iota //语音类型
	FileType_Image                 // 图片类型
	FileType_File                  // 文件类型
	FileType_Video                 //视频类型
)

//func ValidateFileType(fileType FileType, fileName string) bool {
//	switch fileType {
//	case FileType_Voice:
//		return strings.HasSuffix(fileName, ".mp3")
//	case FileType_Image:
//		return strings.HasSuffix(fileName, ".jpg") || strings.HasSuffix(fileName, ".png")
//	case FileType_File:
//		return strings.HasSuffix(fileName, ".doc") || strings.HasSuffix(fileName, ".pdf")
//	case FileType_Video:
//		return strings.HasSuffix(fileName, ".mp4") || strings.HasSuffix(fileName, ".mov")
//	default:
//		return false
//	}
//}

type CompleteUploadRequest struct {
	UploadId string   `json:"upload_id" binding:"required"`
	FileName string   `json:"file_name" binding:"required"`
	Key      string   `json:"key" binding:"required"`
	Type     FileType `json:"type" binding:"required"`
}

type AbortUploadRequest struct {
	UploadId string `json:"upload_id" binding:"required"`
	Key      string `json:"key" binding:"required"`
}
