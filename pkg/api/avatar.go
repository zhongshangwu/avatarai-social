package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type GetAvatarProfileResponse struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Banner      string `json:"banner,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

type UpdateProfileRequest struct {
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	AvatarBase64 string `json:"avatarBase64"`
	BannerBase64 string `json:"bannerBase64,omitempty"`
}

// type UpdateProfileResponse struct {
// 	Success bool `json:"success"`
// 	Profile struct {
// 		Did         string `json:"did"`
// 		Handle      string `json:"handle"`
// 		DisplayName string `json:"displayName"`
// 		Description string `json:"description"`
// 		Avatar      string `json:"avatar"`
// 		Banner      string `json:"banner,omitempty"`
// 		CreatedAt   string `json:"createdAt,omitempty"`
// 	}
// 	Atproto struct {
// 		URI string `json:"uri"`
// 		CID string `json:"cid"`
// 	}
// }

// HandleAvatarProfile 简单地从数据库获取个人资料，并将 CID 转换为 URL
func (a *AvatarAIAPI) HandleAvatarProfile(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	if avatar == nil {
		return ac.Redirect(http.StatusFound, "/api/oauth/login")
	}

	// 从数据库获取个人资料
	var profileData database.Avatar
	if err := a.metaStore.DB.Where("did = ?", avatar.Did).First(&profileData).Error; err != nil {
		return ac.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取个人资料失败: " + err.Error(),
		})
	}

	avatarURL := ""
	if profileData.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			profileData.Did,
			profileData.AvatarCID)
	}

	// 构建响应
	response := GetAvatarProfileResponse{
		Did:         profileData.Did,
		Handle:      profileData.Handle,
		DisplayName: profileData.DisplayName,
		Description: profileData.Description,
		Avatar:      avatarURL,
		CreatedAt:   profileData.CreatedAt.Format(time.RFC3339),
	}

	return ac.JSON(http.StatusOK, response)
}

func (a *AvatarAIAPI) HandleUpdateAvatarProfile(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	oauthSession := ac.OauthSession
	if avatar == nil {
		return ac.JSON(http.StatusUnauthorized, map[string]string{"error": "未授权访问"})
	}

	// 1. 解析请求数据
	var updateReq UpdateProfileRequest
	if err := c.Bind(&updateReq); err != nil {
		return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的请求数据"})
	}

	// 2. 准备更新本地数据库和 atproto PDS 记录
	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	// 创建 vtri.AvatarProfile 记录
	profile := &vtri.AvatarProfile{
		LexiconTypeID: "app.vtri.avatar.profile",
	}

	// 设置基本信息
	did := avatar.Did
	handle := avatar.Handle

	profile.Did = &did
	profile.Handle = &handle

	if updateReq.DisplayName != "" {
		profile.DisplayName = &updateReq.DisplayName
	}

	if updateReq.Description != "" {
		profile.Description = &updateReq.Description
	}

	// 设置创建时间
	createdAt := time.Now().Format(time.RFC3339)
	profile.CreatedAt = &createdAt

	// 为数据库准备更新
	updates := map[string]interface{}{}
	if updateReq.DisplayName != "" {
		updates["display_name"] = updateReq.DisplayName
	}

	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}

	// 3. 如果有头像图片（Base64 格式），上传到 PDS
	if updateReq.AvatarBase64 != "" {
		// 去掉可能的 data:image/xxx;base64, 前缀
		base64Data := updateReq.AvatarBase64
		if idx := strings.Index(base64Data, ","); idx != -1 {
			base64Data = base64Data[idx+1:]
		}

		// 解码 Base64 数据
		imgData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的头像图片数据"})
		}

		// 检测 MIME 类型
		mimeType := http.DetectContentType(imgData)
		if !strings.HasPrefix(mimeType, "image/") {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "上传的文件不是有效的图像"})
		}

		// 上传到 PDS
		var uploadResult struct {
			Blob *util.LexBlob `json:"blob"`
		}

		// 打印图像大小
		log.Println("imgData size", len(imgData))

		err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
			nil, bytes.NewReader(imgData), &uploadResult)

		if err != nil {
			return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "上传头像到 PDS 失败: " + err.Error()})
		}

		// 设置头像
		log.Println("uploadResult.Blob", uploadResult.Blob)
		if uploadResult.Blob != nil {
			profile.Avatar = uploadResult.Blob
			updates["avatar_cid"] = uploadResult.Blob.Ref.String()
		}
	}

	// 4. 如果有背景图片（Base64 格式），上传到 PDS
	if updateReq.BannerBase64 != "" {
		// 去掉可能的 data:image/xxx;base64, 前缀
		base64Data := updateReq.BannerBase64
		if idx := strings.Index(base64Data, ","); idx != -1 {
			base64Data = base64Data[idx+1:]
		}

		// 解码 Base64 数据
		imgData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的背景图片数据"})
		}

		// 检测 MIME 类型
		mimeType := http.DetectContentType(imgData)
		if !strings.HasPrefix(mimeType, "image/") {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "上传的文件不是有效的图像"})
		}

		// 上传到 PDS
		var uploadResult struct {
			Blob *util.LexBlob `json:"blob"`
		}

		err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
			nil, bytes.NewReader(imgData), &uploadResult)

		if err != nil {
			return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "上传背景图到 PDS 失败: " + err.Error()})
		}

		// 设置背景图
		if uploadResult.Blob != nil {
			profile.Banner = uploadResult.Blob
		}

	}

	// 5. 更新 PDS 上的个人资料
	putRecordParams := map[string]interface{}{
		"repo":       avatar.Did,
		"collection": "app.vtri.avatar.profile",
		"rkey":       "self",
		"record":     profile,
	}

	var putResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	err := xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, "application/json", "com.atproto.repo.putRecord",
		nil, putRecordParams, &putResult)

	if err != nil {
		return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "更新 PDS 个人资料失败: " + err.Error()})
	}

	// 6. 更新本地数据库
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := a.metaStore.DB.Model(&database.Avatar{}).Where("did = ?", avatar.Did).Updates(updates).Error; err != nil {
			// 数据库更新失败，但 PDS 更新已成功
			return ac.JSON(http.StatusOK, map[string]interface{}{
				"success": true,
				"warning": "PDS 个人资料已更新，但本地数据库更新失败",
				"pds":     putResult,
			})
		}

		// 更新内存中的avatar对象
		if displayName, ok := updates["display_name"].(string); ok && displayName != "" {
			avatar.DisplayName = displayName
		}

		if description, ok := updates["description"].(string); ok && description != "" {
			avatar.Description = description
		}

		if avatarCID, ok := updates["avatar_cid"].(string); ok && avatarCID != "" {
			avatar.AvatarCID = avatarCID
		}
	}

	// 构建头像URL
	avatarURL := ""
	if avatar.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			avatar.Did,
			avatar.AvatarCID)
	}

	response := GetAvatarProfileResponse{
		Did:         avatar.Did,
		Handle:      avatar.Handle,
		DisplayName: avatar.DisplayName,
		Description: avatar.Description,
		Avatar:      avatarURL,
		Banner:      "",
		CreatedAt:   "",
	}

	return ac.JSON(http.StatusOK, response)
}
