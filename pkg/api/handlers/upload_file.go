package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/types"
)

type BlobHandler struct {
	config      *config.SocialConfig
	metaStore   *repositories.MetaStore
	fileService *services.FileService
}

type UploadFileResponse struct {
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

func NewBlobHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *BlobHandler {
	return &BlobHandler{
		config:      config,
		metaStore:   metaStore,
		fileService: services.NewFileService(metaStore),
	}
}

func (h *BlobHandler) UploadBlobHandler(c *types.APIContext) error {
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "未找到上传文件: "+err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "无法打开上传文件: "+err.Error())
	}
	defer src.Close()

	uploadFile, err := h.fileService.UploadFile(
		c.Request().Context(),
		c.OauthSession.Did,
		c.OauthSession,
		src,
		file.Filename,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "上传文件失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, &UploadFileResponse{
		ID:        uploadFile.ID,
		Size:      uploadFile.Size,
		Filename:  uploadFile.Filename,
		Extension: uploadFile.Extension,
		MimeType:  uploadFile.MimeType,
		CID:       uploadFile.CID,
		URL:       uploadFile.URL,
		CreatedBy: uploadFile.CreatedBy,
		CreatedAt: uploadFile.CreatedAt,
	})
}

func (h *BlobHandler) GetFileHandler(c *types.APIContext) error {
	fileID := c.Param("id")
	if fileID == "" {
		return c.InvalidRequest("file_id is required", "file_id is required")
	}

	file, err := h.fileService.GetFile(c.Request().Context(), fileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "文件不存在: "+err.Error())
	}

	return c.JSON(http.StatusOK, &UploadFileResponse{
		ID:        file.ID,
		Size:      file.Size,
		Filename:  file.Filename,
		Extension: file.Extension,
		MimeType:  file.MimeType,
		CID:       file.CID,
		URL:       file.URL,
		CreatedBy: file.CreatedBy,
		CreatedAt: file.CreatedAt,
	})
}
