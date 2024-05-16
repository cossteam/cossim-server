package storage

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/storage/api/http/v1"
	"github.com/cossim/coss-server/internal/storage/domain/entity"
	"github.com/cossim/coss-server/pkg/code"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"mime/multipart"
	"strings"
)

type StorageService interface {
	Upload(ctx context.Context, userID string, file *multipart.FileHeader, _Type int) (*v1.UploadFileResponse, error)
	GetFileInfo(ctx context.Context, id string) (*entity.File, error)
	DeleteFile(ctx context.Context, id string) error
	GetMultipartUploadKey(ctx context.Context, fileName string, _Type int) (*v1.GetMultipartUploadKeyResponse, error)
	UploadMultipart(ctx context.Context, key string, uploadId string, partNumber int, reader io.Reader, size int64) error
	CompleteMultipartUpload(ctx context.Context, req *v1.CompleteUploadRequest) (string, error)
	AbortMultipartUpload(ctx context.Context, key string, uploadId string) error
}

func (s *ServiceImpl) Upload(ctx context.Context, userID string, file *multipart.FileHeader, _Type int) (*v1.UploadFileResponse, error) {
	fileObj, err := file.Open()
	if err != nil {

		return nil, err
	}

	bucket, err := myminio.GetBucketName(_Type)
	if err != nil {
		return nil, err
	}
	lastDotIndex := strings.LastIndex(file.Filename, ".")
	fileExtension := ""

	if lastDotIndex == -1 || lastDotIndex == len(file.Filename)-1 {
		fileExtension = ""
	} else {
		fileExtension = file.Filename[strings.LastIndex(file.Filename, "."):]
	}

	opt := s.GetContentTypeOption(fileExtension)

	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+fileExtension)

	fmt.Println("s.sp =>", s.sp)
	_, err = s.sp.Upload(ctx, key, fileObj, file.Size, opt)
	if err != nil {
		return nil, err

	}
	_, err = s.sp.GetUrl(context.Background(), key)
	if err != nil {
		return nil, err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, key)
	if s.ac.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return nil, err
		}
	}

	//headerUrl.Host = gatewayAddress
	//if !systemEnableSSL {
	//	headerUrl.Host = gatewayAddress + ":" + gatewayPort
	//}
	//headerUrl.Path = downloadURL + headerUrl.Path
	//
	//aUrl := headerUrl.String()
	//if systemEnableSSL {
	//	aUrl, err = httputil.ConvertToHttps(headerUrl.String())
	//	if err != nil {
	//		h.logger.Error("上传失败", zap.Error(err))
	//		response.SetFail(c, "上传失败", nil)
	//		return
	//	}
	//}

	err = s.sd.Upload(context.Background(), &entity.File{
		Owner: userID,
		Name:  file.Filename,
		Path:  key,
		Type:  entity.FileType(_Type),
		Size:  uint64(file.Size),
	})
	if err != nil {
		return nil, err
	}

	return &v1.UploadFileResponse{
		FileId: fileID,
		Url:    aUrl,
	}, nil
}

func (s *ServiceImpl) GetFileInfo(ctx context.Context, id string) (*entity.File, error) {
	file, err := s.sd.GetFileInfo(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Code(code.StorageErrGetFileInfoFailed.Code()), err.Error())
	}

	//URL := file.Url
	//if strings.Contains(URL, "http://minio:9000") {
	//	URL = strings.Replace(URL, "http://minio:9000", "http://gateway:8080/api/v1/storage/files", 1)
	//}
	//file.Url = URL

	return file, nil
}

func (s *ServiceImpl) DeleteFile(ctx context.Context, fileId string) error {
	resp, err := s.sd.GetFileInfo(ctx, fileId)
	if err != nil {
		return err
	}

	if err = s.sp.Delete(ctx, resp.Path); err != nil {
		return err
	}

	err = s.sd.Delete(ctx, fileId)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServiceImpl) GetMultipartUploadKey(ctx context.Context, fileName string, _Type int) (*v1.GetMultipartUploadKeyResponse, error) {
	// 获取桶名称
	bucket, err := myminio.GetBucketName(_Type)
	if err != nil {
		return nil, err
	}

	// 生成文件ID和文件扩展名
	lastDotIndex := strings.LastIndex(fileName, ".")
	fileExtension := ""
	if lastDotIndex == -1 || lastDotIndex == len(fileName)-1 {
		fileExtension = ""
	} else {
		fileExtension = fileName[lastDotIndex:]
	}
	fileID := uuid.New().String()

	// 生成对象键
	key := myminio.GenKey(bucket, fileID+fileExtension)

	multipartUpload, err := s.sp.NewMultipartUpload(ctx, key, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &v1.GetMultipartUploadKeyResponse{
		UploadId: multipartUpload,
		Type:     _Type,
		Key:      key,
	}, nil
}

func (s *ServiceImpl) UploadMultipart(ctx context.Context, key string, uploadId string, partNumber int, reader io.Reader, size int64) error {

	err := s.sp.UploadPart(ctx, key, uploadId, partNumber, reader, size, minio.PutObjectPartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceImpl) CompleteMultipartUpload(ctx context.Context, req *v1.CompleteUploadRequest) (string, error) {

	_, err := s.sp.CompleteMultipartUpload(context.Background(), req.Key, req.UploadId)
	if err != nil {
		return "", err
	}
	info, err := s.sp.GetObjectInfo(context.Background(), req.Key, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	err = s.sd.Upload(context.Background(), &entity.File{
		Owner: "userID",
		Name:  req.FileName,
		Path:  req.Key,
		Type:  entity.FileType(req.Type),
		Size:  uint64(info.Size),
	})
	if err != nil {
		return "", err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, req.Key)
	if s.ac.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return "", err
		}
	}

	return aUrl, nil
}

func (s *ServiceImpl) AbortMultipartUpload(ctx context.Context, key string, uploadId string) error {
	return s.sp.AbortMultipartUpload(ctx, key, uploadId)
}

func (s *ServiceImpl) GetContentTypeOption(fileExt string) minio.PutObjectOptions {
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
