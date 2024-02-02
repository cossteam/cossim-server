package encryption_test

import (
	"github.com/wneessen/go-mail"
	"log"
	"testing"
)

var colors = []uint32{
	0xff6200, 0x42c58e, 0x5a8de1, 0x785fe0,
}

func TestEncryption(t *testing.T) {
	m := mail.NewMsg()
	if err := m.From("2318266924@qq.com"); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := m.To("2622788078@qq.com"); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}
	m.Subject("老铁拉屎没纸")
	m.SetBodyString(mail.TypeTextPlain, "Do you like this mail? I certainly do!")

	c, err := mail.NewClient("smtp.qq.com", mail.WithPort(25), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername("2318266924@qq.com"), mail.WithPassword("zjnudhwoiuknecgh"))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}
	if err := c.DialAndSend(m); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}
}

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
