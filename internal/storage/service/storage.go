package service

import (
	"context"
	"fmt"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	"github.com/cossim/coss-server/internal/storage/api/http/model"
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

func (s *Service) Upload(ctx context.Context, userID string, file *multipart.FileHeader, _Type int) (*model.UploadFileResponse, error) {
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

	opt := model.GetContentTypeOption(fileExtension)

	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+fileExtension)
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

	_, err = s.storageService.Upload(context.Background(), &storagev1.UploadRequest{
		UserID:   userID,
		FileName: file.Filename,
		Path:     key,
		Url:      aUrl,
		Type:     storagev1.FileType(_Type),
		Size:     uint64(file.Size),
	})
	if err != nil {
		return nil, err
	}

	return &model.UploadFileResponse{
		FileId: fileID,
		Url:    aUrl,
	}, nil
}

func (s *Service) GetFileInfo(ctx context.Context, request *model.GetFileInfoRequest) (interface{}, error) {
	file, err := s.storageService.GetFileInfo(ctx, &storagev1.GetFileInfoRequest{
		FileID: request.FileId,
	})
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

func (s *Service) DeleteFile(ctx context.Context, fileId string) error {
	resp, err := s.storageService.GetFileInfo(ctx, &storagev1.GetFileInfoRequest{FileID: fileId})
	if err != nil {
		return err
	}

	if err = s.sp.Delete(ctx, resp.Path); err != nil {
		return err
	}

	_, err = s.storageService.Delete(ctx, &storagev1.DeleteRequest{FileID: fileId})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetMultipartUploadKey(ctx context.Context, request *model.GetMultipartUploadKeyRequest) (*model.GetMultipartUploadKeyResponse, error) {
	// 获取桶名称
	bucket, err := myminio.GetBucketName(int(request.Type))
	if err != nil {
		return nil, err
	}

	// 生成文件ID和文件扩展名
	lastDotIndex := strings.LastIndex(request.FileName, ".")
	fileExtension := ""
	if lastDotIndex == -1 || lastDotIndex == len(request.FileName)-1 {
		fileExtension = ""
	} else {
		fileExtension = request.FileName[lastDotIndex:]
	}
	fileID := uuid.New().String()

	// 生成对象键
	key := myminio.GenKey(bucket, fileID+fileExtension)

	multipartUpload, err := s.sp.NewMultipartUpload(ctx, key, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &model.GetMultipartUploadKeyResponse{
		UploadId: multipartUpload,
		Type:     request.Type,
		Key:      key,
	}, nil
}

func (s *Service) UploadMultipart(ctx context.Context, key string, uploadId string, partNumber int, reader io.Reader, size int64) error {

	err := s.sp.UploadPart(ctx, key, uploadId, partNumber, reader, size, minio.PutObjectPartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CompleteMultipartUpload(ctx context.Context, req *model.CompleteUploadRequest) (string, error) {

	headerUrl, err := s.sp.CompleteMultipartUpload(context.Background(), req.Key, req.UploadId)
	if err != nil {
		return "", err
	}
	info, err := s.sp.GetObjectInfo(context.Background(), req.Key, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	_, err = s.storageService.Upload(context.Background(), &storagev1.UploadRequest{
		UserID:   "userID",
		FileName: req.FileName,
		Path:     req.Key,
		Url:      headerUrl.String(),
		Type:     storagev1.FileType(req.Type),
		Size:     uint64(info.Size),
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
	//headerUrl.Host = gatewayAddress + ":" + gatewayPort
	//headerUrl.Path = downloadURL + headerUrl.Path
}

func (s *Service) AbortMultipartUpload(ctx context.Context, key string, uploadId string) error {
	return s.sp.AbortMultipartUpload(ctx, key, uploadId)
}
