package http

import (
	"context"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	"github.com/cossim/coss-server/internal/storage/api/http/model"
	http2 "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// upload
// @Summary 上传文件
// @Description 上传文件
// @Tags Storage
// @param file formData file true "文件"
// @param type formData integer false "文件类型(0:音频，1:图片，2:文件，3:视频)"
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /storage/files [post]
func (h *Handler) upload(c *gin.Context) {
	userID, err := http2.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	// 获取表单中的整数字段，如果字段不存在或无法解析为整数，则使用默认值 0
	value := c.PostForm("type")
	if value == "" {
		value = "2"
	}

	_Type, err := strconv.Atoi(value)
	if err != nil {
		h.logger.Error("type解析失败", zap.Error(err))
		response.SetFail(c, "文件类型解析失败", nil)
		return
	}

	// 文件大小限制
	maxFileSize := storagev1.GetMaxFileSize(storagev1.FileType(_Type))
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("上传失败", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}

	if file.Size > maxFileSize {
		h.logger.Error("文件大小超过限制", zap.Error(err))
		response.SetFail(c, "文件大小超过限制", nil)
		return
	}

	resp, err := h.svc.Upload(c, userID, file, _Type)
	if err != nil {
		return
	}

	response.SetSuccess(c, "上传成功", resp)
}

// download
// @Summary 下载文件
// @Description 下载文件
// @Tags Storage
// @param id path string true "文件id"
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /storage/files/download/:type/:id [get]
func (h *Handler) download(c *gin.Context) {
	targetURL := "http://" + h.minioAddr
	URL := c.Request.URL.String()
	if strings.Contains(URL, downloadURL) {
		URL = strings.Replace(URL, downloadURL, "", 1)
	}
	targetURL += URL

	// 创建一个代理请求
	proxyReq, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to create proxy request", zap.Error(err))
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
		h.logger.Error("Failed to fetch response from service", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from service"})
		return
	}
	defer resp.Body.Close()

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
// @Tags Storage
// @param id path string true "文件id"
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /storage/files/:id [get]
func (h *Handler) getFileInfo(c *gin.Context) {
	//fileID := c.Query("file_id")
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	info, err := h.svc.GetFileInfo(context.Background(), &model.GetFileInfoRequest{FileId: fileID})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取文件信息成功", info)
}

// deleteFile
// @Summary 删除文件
// @Description 删除文件
// @param id path string true "文件id"
// @Produce  json
// @Tags Storage
// @Success		200 {object} model.Response{}
// @Router /storage/files/:id [delete]
func (h *Handler) deleteFile(c *gin.Context) {
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

	err := h.svc.DeleteFile(context.Background(), fileID)
	if err != nil {
		return
	}

	response.SetSuccess(c, "success", nil)
}

// @Summary 生成分片上传id
// @Description 生成分片上传id
// @Tags Storage
// @Produce  json
// @param file_name query string true "文件名"
// @param type query integer false "文件类型(0:音频，1:图片，2:文件，3:视频)"
// @Success		200 {object} model.Response{}
// @Router /storage/files/multipart/key [get]
func (h *Handler) getMultipartKey(c *gin.Context) {
	fileName := c.Query("file_name")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_name is required"})
		return
	}
	fileType := c.Query("type")
	if fileType == "" {
		fileType = "2"
	}
	t, err := strconv.Atoi(fileType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}

	resp, err := h.svc.GetMultipartUploadKey(context.Background(), &model.GetMultipartUploadKeyRequest{
		FileName: fileName,
		Type:     model.FileType(t),
	})
	if err != nil {
		return
	}

	response.SetSuccess(c, "获取文件信息成功", resp)
}

// @Summary 上传分片
// @Description 上传分片
// @Tags Storage
// @param file formData file true "本次分片"
// @param upload_id formData string true "上传id"
// @param part_number formData integer true "本次分片序号"
// @param key formData string true "文件唯一key"
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /storage/files/multipart/upload [post]
func (h *Handler) uploadMultipart(c *gin.Context) {
	//单次分片限制100m
	maxFileSize := 100 * 1024 * 1024
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("上传失败", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}

	if file.Size > int64(maxFileSize) {
		h.logger.Error("文件大小超过限制", zap.Error(err))
		response.SetFail(c, "文件大小超过限制", nil)
		return
	}

	uploadId := c.PostForm("upload_id")
	if uploadId == "" {
		h.logger.Error("upload_id is required", zap.Error(err))
		response.SetFail(c, "upload_id is required", nil)
		return
	}
	number := c.PostForm("part_number")
	if uploadId == "" {
		h.logger.Error("part_number is required", zap.Error(err))
		response.SetFail(c, "part_number is required", nil)
		return
	}
	partNumber, err := strconv.Atoi(number)
	if err != nil {
		h.logger.Error("part_number解析失败", zap.Error(err))
		response.SetFail(c, "part_number解析失败", nil)
		return
	}

	key := c.PostForm("key")
	if key == "" {
		h.logger.Error("key is required", zap.Error(err))
		response.SetFail(c, "key is required", nil)
		return
	}

	fileObj, err := file.Open()
	if err != nil {
		h.logger.Error("上传失败", zap.Error(err))
		response.SetFail(c, "上传失败", nil)
		return
	}

	err = h.svc.UploadMultipart(c, key, uploadId, partNumber, fileObj, file.Size)
	if err != nil {
		h.logger.Error("上传失败", zap.Error(err))
		response.SetFail(c, "上传失败", nil)
		return
	}

	response.SetSuccess(c, "分片上传成功", nil)
}

// @Summary 完成分片上传
// @Description 完成分片上传
// @Tags Storage
// @Produce  json
// @Accept  json
// @param request body model.CompleteUploadRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /storage/files/multipart/complete [post]
func (h *Handler) completeUploadMultipart(c *gin.Context) {
	req := new(model.CompleteUploadRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	resp, err := h.svc.CompleteMultipartUpload(c, req)
	if err != nil {
		return
	}

	response.SetSuccess(c, "上传成功", gin.H{"file_url": resp})
}

// @Summary 清除文件分片(用于中断上传)
// @Description 清除文件分片
// @Tags Storage
// @Produce  json
// @Accept  json
// @param request body model.AbortUploadRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /storage/files/multipart/abort [post]
func (h *Handler) abortUploadMultipart(c *gin.Context) {
	req := new(model.AbortUploadRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	err := h.svc.AbortMultipartUpload(context.Background(), req.Key, req.UploadId)
	if err != nil {
		h.logger.Error("清理失败", zap.Error(err))
		response.SetFail(c, "清理失败", nil)
		return
	}

	response.SetSuccess(c, "清理成功", nil)
}
