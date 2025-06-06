package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type UploadBlobResponse struct {
	ID        string `json:"id"`
	Size      int64  `json:"size"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
	CID       string `json:"cid"`
	URL       string `json:"url"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

func generateFileID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func detectMimeTypeAndExtension(filename string, fileBytes []byte) (string, string) {
	// 1. 先从文件内容检测MIME类型
	detectedMimeType := http.DetectContentType(fileBytes)

	// 2. 从文件名获取扩展名
	ext := strings.ToLower(filepath.Ext(filename))

	// 3. 尝试从扩展名获取更准确的MIME类型
	if ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			detectedMimeType = mimeType
		}
	}

	// 4. 如果没有扩展名，尝试从MIME类型推断扩展名
	if ext == "" {
		switch detectedMimeType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		case "video/mp4":
			ext = ".mp4"
		case "video/webm":
			ext = ".webm"
		case "video/quicktime":
			ext = ".mov"
		case "audio/mpeg":
			ext = ".mp3"
		case "audio/wav":
			ext = ".wav"
		case "application/pdf":
			ext = ".pdf"
		case "text/plain":
			ext = ".txt"
		default:
			// 根据主要类型设置默认扩展名
			if strings.HasPrefix(detectedMimeType, "image/") {
				ext = ".jpg"
			} else if strings.HasPrefix(detectedMimeType, "video/") {
				ext = ".mp4"
			} else if strings.HasPrefix(detectedMimeType, "audio/") {
				ext = ".mp3"
			} else {
				ext = ".bin"
			}
		}
	}

	// 移除扩展名的点号
	if ext != "" && ext[0] == '.' {
		ext = ext[1:]
	}

	return detectedMimeType, ext
}

func (a *AvatarAIAPI) UploadBlobHandler(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession
	avatar := ac.Avatar

	if avatar == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "未授权访问")
	}

	// 1. 获取请求里的 file 对象
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "未找到上传文件: "+err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "无法打开上传文件: "+err.Error())
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "读取文件失败: "+err.Error())
	}

	// 检查文件大小限制（例如：50MB）
	const maxFileSize = 50 * 1024 * 1024
	if len(fileBytes) > maxFileSize {
		return echo.NewHTTPError(http.StatusBadRequest, "文件大小超过限制")
	}

	// 2. 获取文件名，mimetype，和 extension，智能和鲁棒处理
	filename := file.Filename
	if filename == "" {
		filename = "unnamed_file"
	}

	mimeType, extension := detectMimeTypeAndExtension(filename, fileBytes)

	// 验证文件类型是否被支持
	supportedTypes := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"video/mp4", "video/webm", "video/quicktime",
		"audio/mpeg", "audio/wav",
		"application/pdf", "text/plain",
	}

	isSupported := false
	for _, supportedType := range supportedTypes {
		if mimeType == supportedType || strings.HasPrefix(mimeType, supportedType) {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return echo.NewHTTPError(http.StatusBadRequest, "不支持的文件类型: "+mimeType)
	}

	log.Printf("文件信息: 名称=%s, 大小=%d, MIME=%s, 扩展名=%s",
		filename, len(fileBytes), mimeType, extension)

	// 3. 通过 atp 存储到 pds 中
	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	var uploadResult struct {
		Blob *util.LexBlob `json:"blob"`
	}

	err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
		nil, bytes.NewReader(fileBytes), &uploadResult)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "上传文件到 PDS 失败: "+err.Error())
	}

	if uploadResult.Blob == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "PDS 上传结果为空")
	}

	blob := uploadResult.Blob

	// 生成访问URL
	url := fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s",
		oauthSession.PdsUrl, oauthSession.Did, blob.Ref.String())

	// 4. 存储到 upload file 数据库
	fileID := generateFileID()
	uploadFile := &database.UploadFile{
		ID:        fileID,
		Size:      int64(len(fileBytes)),
		Filename:  filename,
		Extension: extension,
		MimeType:  mimeType,
		CID:       blob.Ref.String(),
		CreatedBy: avatar.Did,
	}

	if err := database.CreateUploadFile(a.metaStore.DB, uploadFile); err != nil {
		log.Printf("保存文件记录到数据库失败: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "保存文件记录到数据库失败: "+err.Error())
	}

	// 5. 返回响应
	response := UploadBlobResponse{
		ID:        uploadFile.ID,
		Size:      uploadFile.Size,
		Filename:  uploadFile.Filename,
		Extension: uploadFile.Extension,
		MimeType:  uploadFile.MimeType,
		CID:       uploadFile.CID,
		URL:       url,
		CreatedBy: uploadFile.CreatedBy,
		CreatedAt: uploadFile.CreatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

func (a *AvatarAIAPI) GetUploadFilesHandler(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar

	if avatar == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "未授权访问")
	}

	// 获取查询参数
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 20 // 默认限制
	offset := 0 // 默认偏移

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 从数据库获取文件列表
	files, err := database.GetUploadFilesByCreator(a.metaStore.DB, avatar.Did, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取文件列表失败: "+err.Error())
	}

	// 为每个文件生成URL
	oauthSession := ac.OauthSession
	for _, file := range files {
		file.URL = fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s",
			oauthSession.PdsUrl, oauthSession.Did, file.CID)
	}

	// 获取统计信息
	stats, err := database.GetUploadFileStats(a.metaStore.DB, avatar.Did)
	if err != nil {
		log.Printf("获取文件统计信息失败: %v", err)
		stats = map[string]interface{}{
			"total_count": 0,
			"total_size":  0,
		}
	}

	response := map[string]interface{}{
		"files": files,
		"stats": stats,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(files),
		},
	}

	return c.JSON(http.StatusOK, response)
}
