package encryption_test

//import (
//	"fmt"
//	"github.com/sony/sonyflake"
//)
//
//func GenCossID() (string, error) {
//	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
//	id, err := flake.NextID()
//	if err != nil {
//		return "", err
//	}
//	return fmt.Sprintf("coss_id_%x", id), nil
//}

//func TestEncryption(t *testing.T) {
//	//fmt.Println("Generated ID:", id)
//}

//func min(a, b int64) int64 {
//	if a < b {
//		return a
//	}
//	return b
//}
//func TestMulpUpload(t *testing.T) {
//
//	var err error
//
//	sp, err := myminio.NewMinIOStorage("127.0.0.1:9000", "E003IlrA7X83C5hmS6oV", "Ns7lKyU9JWPpQ21L6z6SMOMzx1ZgX2dxFQTRKZcs", false)
//	if err != nil {
//
//		panic(err)
//	}
//
//	//打开文件
//	file, err := os.Open("ubuntu-22.04.3-desktop-amd64.iso")
//	if err != nil {
//		return
//	}
//	defer file.Close()
//
//	// 获取文件长度
//	fileInfo, err := file.Stat()
//	if err != nil {
//		return
//	}
//	fileSize := fileInfo.Size()
//	chunkSize := int64(1024 * 1024 * 1024) // 1GB
//
//	// 计算需要的分片数
//	numParts := (fileSize + chunkSize - 1) / chunkSize
//
//	// 获取桶名称
//	bucket, err := myminio.GetBucketName(3)
//	if err != nil {
//		panic(err)
//	}
//
//	// 生成文件ID和文件扩展名
//	lastDotIndex := strings.LastIndex(file.Name(), ".")
//	fileExtension := ""
//	if lastDotIndex == -1 || lastDotIndex == len(file.Name())-1 {
//		fileExtension = ""
//	} else {
//		fileExtension = file.Name()[lastDotIndex:]
//	}
//	fileID := uuid.New().String()
//
//	// 生成对象键
//	key := myminio.GenKey(bucket, fileID+fileExtension)
//
//	// 启动分片上传
//	upload, err := sp.NewMultipartUpload(context.Background(), key, minio.PutObjectOptions{})
//	if err != nil {
//		return
//	}
//
//	// 分片上传
//	for i := int64(0); i < numParts; i++ {
//		// 创建分片大小的缓冲区
//		partSize := min(chunkSize, fileSize-i*chunkSize)
//		part := make([]byte, partSize)
//
//		// 从文件中读取分片数据
//		_, err = file.Read(part)
//		if err != nil {
//			return
//		}
//
//		// 上传分片
//		err = sp.UploadPart(context.Background(), key, upload, int(i+1), bytes.NewReader(part), partSize, minio.PutObjectPartOptions{})
//		if err != nil {
//			return
//		}
//	}
//	fmt.Println("上传完毕。。。")
//	re, err := sp.CompleteMultipartUpload(context.Background(), key, upload)
//	if err != nil {
//		return
//	}
//	fmt.Println("string =>", re.String())
//}
