package minio

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/storage"
	storev1 "github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/policy"
	"io"
	"net/url"
	"strings"
	"time"
)

// FileBucket 是文件桶的默认名称
const FileBucket = "file"

// AudioBucket 是音频桶的默认名称
const AudioBucket = "audio"

var BucketList = map[storev1.FileType]string{
	//storev1.FileType_Text:  FileBucket,
	storev1.FileType_Voice: AudioBucket,
	storev1.FileType_Image: FileBucket,
	storev1.FileType_File:  FileBucket,
	storev1.FileType_Video: AudioBucket,
	//storev1.FileType_EMOJI: FileBucket,
	//storev1.FileType_Sticker: FileBucket,
}

func NewMinIOStorage(endpoint, accessKey, secretKey string, useSSL bool, opts ...func(*MinIOStorage)) (storage.StorageProvider, error) {
	c := &MinIOStorage{
		Endpoint:         endpoint,
		AccessKey:        accessKey,
		SecretKey:        secretKey,
		UseSSL:           useSSL,
		PresignedExpires: time.Hour * 24,
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

	for _, v := range []string{FileBucket, AudioBucket} {
		if err = c.CreateMinoBuket(v, 0); err != nil {
			panic(err)
		}
	}

	return c, nil
}

// MinIOStorage 使用 MinIO 客户端实现 StorageProvider 接口
type MinIOStorage struct {
	client           *minio.Client
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
	// 设置存储桶策略
	if err = m.client.SetBucketPolicy(context.Background(), bucketName, string(policy.BucketPolicyReadWrite)); err != nil {
		return err
	}
	// 将存储桶名称与文件类型关联起来
	//m.BucketList[storev1.FileType(fileType)] = bucketName
	fmt.Printf("Successfully created %s bucket for file type %v\n", bucketName, fileType)

	return nil
}
