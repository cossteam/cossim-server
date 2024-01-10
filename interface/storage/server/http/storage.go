package http

import (
	"context"
	conf "github.com/cossim/coss-server/interface/storage/config"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/storage/minio"
	storagev1 "github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// upload
// @Summary 上传文件
// @Description 上传文件
// @param file formData file true "文件"
// @param type formData integer false "文件类型"
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /files [post]
func upload(c *gin.Context) {
	//userID, err := http.ParseTokenReUid(c)
	//if err != nil {
	//	logger.Error("token解析失败", zap.Error(err))
	//	response.Fail(c, "token解析失败", nil)
	//	return
	//}

	// 获取表单中的整数字段，如果字段不存在或无法解析为整数，则使用默认值 0
	value := c.PostForm("type")
	_Type, err := strconv.Atoi(value)
	if err != nil {
		logger.Error("上传失败", zap.Error(err))
		response.Fail(c, "上传失败", nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("上传失败", zap.Error(err))
		response.Fail(c, err.Error(), nil)
		return
	}
	fileObj, err := file.Open()
	if err != nil {
		logger.Error("上传失败", zap.Error(err))
		response.Fail(c, "上传失败", nil)
		return
	}

	bucket, err := minio.GetBucketName(_Type)
	if err != nil {
		logger.Error("上传失败", zap.Error(err))
		response.Fail(c, "上传失败", nil)
		return
	}

	fileID := uuid.New().String()
	key := minio.GenKey(bucket, fileID)

	headerUrl, err := sp.Upload(context.Background(), key, fileObj, file.Size)
	if err != nil {
		logger.Error("上传失败", zap.Error(err))
		response.Fail(c, "上传失败", nil)
		return
	}

	_, err = storageClient.Upload(context.Background(), &storagev1.UploadRequest{
		UserID:   "userID",
		FileName: file.Filename,
		Path:     key,
		Url:      headerUrl.String(),
		Type:     storagev1.FileType(_Type),
		Size:     uint64(file.Size),
	})
	if err != nil {
		logger.Error("上传失败", zap.Error(err))
		response.Fail(c, "上传失败", nil)
		return
	}

	headerUrl.Host = cfg.Discovers["gateway"].Addr
	headerUrl.Path = downloadURL + headerUrl.Path
	response.Success(c, "上传成功", gin.H{
		"url":     headerUrl.String(),
		"file_id": fileID,
	})
}

// download
// @Summary 下载文件
// @Description 下载文件
// @param id path string true "文件id"
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /files/download/:type/:id [get]
func download(c *gin.Context) {
	targetURL := "http://" + conf.MinioConf.Endpoint
	URL := c.Request.URL.String()
	if strings.Contains(URL, downloadURL) {
		URL = strings.Replace(URL, downloadURL, "", 1)
	}
	targetURL += URL

	// 创建一个代理请求
	proxyReq, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		logger.Error("Failed to create proxy request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	// 添加查询字符串到代理请求的 URL 中
	proxyReq.URL.RawQuery = c.Request.URL.RawQuery

	// 复制请求头信息
	proxyReq.Header = make(http.Header)
	for h, val := range c.Request.Header {
		proxyReq.Header[h] = val
	}

	// 发送代理请求
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		logger.Error("Failed to fetch response from service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from service"})
		return
	}
	defer resp.Body.Close()

	//logger.Info("Received response from service", zap.Any("ResponseHeaders", resp.Header), zap.String("TargetURL", targetURL))

	// 将 BFF 服务的响应返回给客户端
	c.Status(resp.StatusCode)
	for h, val := range resp.Header {
		c.Header(h, val[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// getFileInfo
// @Summary 获取文件信息
// @Description 获取文件信息
// @param id path string true "文件id"
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /files/:id [get]
func getFileInfo(c *gin.Context) {
	//fileID := c.Query("file_id")
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	info, err := storageClient.GetFileInfo(context.Background(), &storagev1.GetFileInfoRequest{FileID: fileID})
	if err != nil {
		c.Error(err)
		return
	}

	URL := info.Url
	if strings.Contains(URL, "http://minio:9000") {
		URL = strings.Replace(URL, "http://minio:9000", "http://gateway:8080/api/v1/storage/files", 1)
	}
	info.Url = URL

	response.SetSuccess(c, "获取文件信息成功", gin.H{"info": info})
}

// deleteFile
// @Summary 删除文件
// @Description 删除文件
// @param id path string true "文件id"
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /files/:id [delete]
func deleteFile(c *gin.Context) {
	//req := &DeleteFileRequest{}
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	return
	//}
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	resp, err := storageClient.GetFileInfo(context.Background(), &storagev1.GetFileInfoRequest{FileID: fileID})
	if err != nil {
		logger.Error("获取文件信息失败", zap.Error(err))
		response.Fail(c, "删除文件失败，请重试", nil)
		//c.Error(err)
		return
	}

	if err = sp.Delete(context.Background(), resp.Path); err != nil {
		logger.Error("oss删除文件失败", zap.Error(err))
		response.Fail(c, "删除文件失败，请重试", nil)
		//c.Error(err)
		return
	}

	_, err = storageClient.Delete(context.Background(), &storagev1.DeleteRequest{FileID: fileID})
	if err != nil {
		response.Fail(c, "删除文件失败，请重试", nil)
		//c.Error(err)
		return
	}

	response.Success(c, "success", nil)
}
