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
	Size      int64  `json:"size"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
	MimeType  string `json:"mimeType"`
	CID       string `json:"cid"`
	URL       string `json:"url"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
}

func NewBlobHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *BlobHandler {
	return &BlobHandler{
		config:      config,
		metaStore:   metaStore,
		fileService: services.NewFileService(config, metaStore),
	}
}

func (h *BlobHandler) UploadFile(c *types.APIContext) error {
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

func (h *BlobHandler) GetFile(c *types.APIContext) error {
	fileCid := c.Param("cid")
	if fileCid == "" {
		return c.InvalidRequest("fileCid is required", "fileCid is required")
	}

	file, err := h.fileService.GetFile(c.Request().Context(), fileCid)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "文件不存在: "+err.Error())
	}

	return c.JSON(http.StatusOK, &UploadFileResponse{
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
