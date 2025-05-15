package api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type UploadMediaResponse struct {
	Blob *util.LexBlob `json:"blob"` // 媒体 CID
	CID  string        `json:"cid"`  // 媒体 CID
	URL  string        `json:"url"`  // 媒体访问 URL
	Type string        `json:"type"` // 媒体类型
}

func (a *AvatarAIAPI) UploadBlobHandler(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession

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

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
	}

	// 调用 ATProto 上传 Blob
	// 注意: 实际项目中需要替换为正确的 repo.uploadBlob 调用
	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	var uploadResult struct {
		Blob *util.LexBlob `json:"blob"`
	}

	// 打印图像大小
	log.Println("imgData size", len(fileBytes))

	err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, contentType, "com.atproto.repo.uploadBlob",
		nil, bytes.NewReader(fileBytes), &uploadResult)

	if err != nil {
		return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "上传头像到 PDS 失败: " + err.Error()})
	}

	blob := uploadResult.Blob
	mediaType := "image"

	if strings.HasPrefix(contentType, "video/") {
		mediaType = "video"
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s",
		oauthSession.PdsUrl, oauthSession.Did, blob.Ref.String())

	// 返回响应
	response := UploadMediaResponse{
		Blob: blob,
		CID:  blob.Ref.String(),
		URL:  url,
		Type: mediaType,
	}
	return c.JSON(http.StatusOK, response)
}
