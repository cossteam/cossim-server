package minio

import (
	"context"
	"fmt"
	storev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FileBucket 是文件桶的默认名称
const FileBucket = "file"

// AudioBucket 是音频桶的默认名称
const AudioBucket = "audio"

// 公开桶
const PublicBucket = "public"

var BucketList = map[storev1.FileType]string{
	//storev1.FileType_Text:  FileBucket,
	storev1.FileType_Voice: AudioBucket,
	storev1.FileType_Image: FileBucket,
	storev1.FileType_File:  FileBucket,
	storev1.FileType_Video: AudioBucket,
	storev1.FileType_Other: PublicBucket,
	//storev1.FileType_EMOJI: FileBucket,
	//storev1.FileType_Sticker: FileBucket,
}

func NewMinIOStorage(endpoint, accessKey, secretKey string, useSSL bool, opts ...func(*MinIOStorage)) (storage.StorageProvider, error) {
	c := &MinIOStorage{
		Endpoint:         endpoint,
		AccessKey:        accessKey,
		SecretKey:        secretKey,
		UseSSL:           useSSL,
		PresignedExpires: time.Hour * 24 * 7,
		BucketList:       BucketList,
	}
	for _, opt := range opts {
		opt(c)
	}
	minioClient, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: c.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	c.client = minioClient

	coreCli, err := minio.NewCore(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: c.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	c.coreClient = coreCli

	for _, v := range []string{FileBucket, AudioBucket, PublicBucket} {
		if err = c.CreateMinoBuket(v, 0); err != nil {
			panic(err)
		}
	}

	return c, nil
}

// MinIOStorage 使用 MinIO 客户端实现 StorageProvider 接口
type MinIOStorage struct {
	client           *minio.Client
	coreClient       *minio.Core
	Endpoint         string
	AccessKey        string
	SecretKey        string
	UseSSL           bool
	PresignedExpires time.Duration
	BucketList       map[storev1.FileType]string // 用于存储文件类型和其对应的存储桶名称
}

func GetBucketName(fileType int) (string, error) {
	bucketName, found := BucketList[storev1.FileType(fileType)]
	if !found {
		return "", fmt.Errorf("bucket not found for file type %d", fileType)
	}
	return bucketName, nil
}

func GenKey(bucketName, objectName string) string {
	return fmt.Sprintf("%s/%s", bucketName, objectName)
}

func ParseKey(key string) (bucketName, objectName string, err error) {
	split := strings.SplitN(key, "/", 2)
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid key format")
	}
	return split[0], split[1], nil
}

func (m *MinIOStorage) Upload(ctx context.Context, key string, reader io.Reader, objectSize int64, opt minio.PutObjectOptions) (*url.URL, error) {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return nil, err
	}

	_, err = m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opt)
	if err != nil {
		return nil, err
	}

	reqParams := make(url.Values)
	//reqParams.Set("response-content-disposition", "inline")

	presignedURL, err := m.client.PresignedGetObject(ctx, bucketName, objectName, m.PresignedExpires, reqParams)
	if err != nil {
		return nil, err
	}

	return presignedURL, nil
}

func (m *MinIOStorage) UploadAvatar(ctx context.Context, key string, reader io.Reader, size int64, opt minio.PutObjectOptions) error {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return err
	}

	_, err = m.client.PutObject(ctx, bucketName, objectName, reader, size, opt)
	if err != nil {
		return err
	}

	return nil
}

func (m *MinIOStorage) GetUrl(ctx context.Context, key string) (string, error) {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return "", err
	}

	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, bucketName, objectName, m.PresignedExpires, reqParams)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), err
}

func (m *MinIOStorage) Delete(ctx context.Context, key string) error {
	bucketName, fileName, err := ParseKey(key)
	if err != nil {
		return err
	}
	if err = m.client.RemoveObject(ctx, bucketName, fileName, minio.RemoveObjectOptions{}); err != nil {
		return err
	}
	return nil
}

// CreateMinoBuket 创建minio 桶
func (m *MinIOStorage) CreateMinoBuket(bucketName string, fileType int) error {
	//fmt.Println("bucketName => ", bucketName)
	// 先检查存储桶是否已经存在
	exists, err := m.client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// 创建存储桶
	if err = m.client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: bucketName}); err != nil {
		return err
	}
	//if bucketName == PublicBucket {
	//	//// 设置存储桶访问策略为公开读
	//	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucketName + `/*"]}]}`
	//
	//	err = m.client.SetBucketPolicy(context.Background(), bucketName, policy)
	//	if err != nil {
	//		return err
	//	}
	//
	//} else {
	//	// 设置存储桶策略
	//	if err = m.client.SetBucketPolicy(context.Background(), bucketName, string(policy.BucketPolicyReadWrite)); err != nil {
	//		return err
	//	}
	//}
	//// 统一设置存储桶访问策略为公开读
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucketName + `/*"]}]}`

	err = m.client.SetBucketPolicy(context.Background(), bucketName, policy)
	if err != nil {
		return err
	}

	// 将存储桶名称与文件类型关联起来
	//m.BucketList[storev1.FileType(fileType)] = bucketName
	fmt.Printf("Successfully created %s bucket for file type %v\n", bucketName, fileType)

	return nil
}

// 发起分片上传请求
func (m *MinIOStorage) NewMultipartUpload(ctx context.Context, key string, opt minio.PutObjectOptions) (string, error) {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return "", err
	}

	upload, err := m.coreClient.NewMultipartUpload(ctx, bucketName, objectName, opt)
	if err != nil {
		return "", err
	}
	return upload, nil
}

func (m *MinIOStorage) GenUploadPartSignedUrl(ctx context.Context, key string, uploadId string, partNumber int, partSize int64) (string, error) {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"uploadId":   []string{uploadId},
		"partNumber": []string{strconv.Itoa(partNumber)},
	}

	object, err := m.coreClient.PresignedGetObject(ctx, bucketName, objectName, m.PresignedExpires, params)
	if err != nil {
		return "", err
	}

	return object.String(), nil
}

func (m *MinIOStorage) CompleteMultipartUpload(ctx context.Context, key string, uploadId string) (*url.URL, error) {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return nil, err
	}
	parts, err := m.coreClient.ListObjectParts(ctx, bucketName, objectName, uploadId, 0, 0)
	if err != nil {
		return nil, err
	}
	list := make([]minio.CompletePart, 0)
	for _, part := range parts.ObjectParts {
		list = append(list, minio.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		})
	}
	//根据PartNumber排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].PartNumber < list[j].PartNumber
	})

	_, err = m.coreClient.CompleteMultipartUpload(ctx, bucketName, objectName, uploadId, list, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	params := url.Values{}

	object, err := m.coreClient.PresignedGetObject(ctx, bucketName, objectName, m.PresignedExpires, params)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (m *MinIOStorage) UploadPart(ctx context.Context, key string, uploadId string, partNumber int, reader io.Reader, size int64, opt minio.PutObjectPartOptions) error {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return err
	}
	// 获取已上传的分片列表
	parts, err := m.coreClient.ListObjectParts(ctx, bucketName, objectName, uploadId, 0, 0)
	if err != nil {
		return err
	}

	// 检查当前分片是否已存在
	for _, part := range parts.ObjectParts {
		if part.PartNumber == partNumber {
			return nil // 当前分片已存在，直接返回
		}
	}

	_, err = m.coreClient.PutObjectPart(ctx, bucketName, objectName, uploadId, partNumber, reader, size, opt)
	if err != nil {
		return err
	}
	return nil
}

// 中止并清理上传的分片
func (m *MinIOStorage) AbortMultipartUpload(ctx context.Context, key string, uploadId string) error {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return err
	}
	return m.coreClient.AbortMultipartUpload(ctx, bucketName, objectName, uploadId)
}

func (m *MinIOStorage) GetObjectInfo(ctx context.Context, key string, opt minio.GetObjectOptions) (minio.ObjectInfo, error) {
	bucketName, objectName, err := ParseKey(key)
	if err != nil {
		return minio.ObjectInfo{}, err
	}
	_, info, _, err := m.coreClient.GetObject(ctx, bucketName, objectName, opt)
	if err != nil {
		return minio.ObjectInfo{}, err
	}
	return info, nil
}
