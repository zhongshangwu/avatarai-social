package api

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/mint"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type AsterProfileResponse struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Banner      string `json:"banner,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

type GetAsterProfileResponse struct {
	Aster       *AsterProfileResponse `json:"aster"`
	Initialized bool                  `json:"initialized"`
}

type UpdateAsterProfileRequest struct {
	DisplayName  string        `json:"displayName"`
	Description  string        `json:"description"`
	AvatarBlob   *util.LexBlob `json:"avatarBlob"`
	AvatarBase64 string        `json:"avatarBase64"`
	BannerBase64 string        `json:"bannerBase64,omitempty"`
}

func (a *AvatarAIAPI) HandleAsterProfile(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	aster, err := database.GetAsterByCreatorDid(a.metaStore.DB, avatar.Did)
	if err != nil {
		if errors.Is(err, database.ErrAsterNotFound) {
			return c.JSON(200, GetAsterProfileResponse{
				Aster:       nil,
				Initialized: false,
			})
		}
		return c.JSON(500, map[string]string{
			"error": "获取Aster信息失败: " + err.Error(),
		})
	}

	// 构建头像URL
	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.CreatorDid,
			aster.AvatarCID)
	}

	// 构建 AsterProfileResponse
	profile := &AsterProfileResponse{
		Did:         aster.Did,
		Handle:      aster.Handle,
		DisplayName: aster.DisplayName,
		Description: aster.Description,
		Avatar:      avatarURL,
		CreatedAt:   aster.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(200, GetAsterProfileResponse{
		Aster:       profile,
		Initialized: true,
	})
}

func (a *AvatarAIAPI) HandleAsterUpdateProfile(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	oauthSession := ac.OauthSession
	if avatar == nil {
		return ac.JSON(http.StatusUnauthorized, map[string]string{"error": "未授权访问"})
	}

	// 1. 获取Aster
	aster, err := database.GetAsterByCreatorDid(a.metaStore.DB, avatar.Did)
	if err != nil {
		if errors.Is(err, database.ErrAsterNotFound) {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "您还没有创建Aster"})
		}
		return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "获取Aster信息失败: " + err.Error()})
	}

	// 2. 解析请求数据
	var updateReq UpdateAsterProfileRequest
	if err := c.Bind(&updateReq); err != nil {
		return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的请求数据"})
	}

	// 3. 准备更新本地数据库和 atproto PDS 记录
	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	// 创建 vtri.AsterProfile 记录
	profile := &vtri.AsterProfile{
		LexiconTypeID: "app.vtri.aster.profile",
	}

	// 设置基本信息
	did := aster.Did
	handle := aster.Handle

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

	// 设置creator引用
	// 创建一个指向创建者的强引用
	creatorUri := fmt.Sprintf("at://%s/app.vtri.avatar.profile/self", avatar.Did)
	profile.Creator = &comatproto.RepoStrongRef{
		LexiconTypeID: "com.atproto.repo.strongRef",
		Uri:           creatorUri,
	}

	// 为数据库准备更新
	updates := map[string]interface{}{}
	if updateReq.DisplayName != "" {
		updates["display_name"] = updateReq.DisplayName
	}

	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}

	// 4. 如果有头像图片（Base64 格式），上传到 PDS
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

		err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
			nil, bytes.NewReader(imgData), &uploadResult)

		if err != nil {
			return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "上传头像到 PDS 失败: " + err.Error()})
		}

		// 设置头像
		if uploadResult.Blob != nil {
			profile.Avatar = uploadResult.Blob
			updates["avatar_cid"] = uploadResult.Blob.Ref.String()
		}
	}

	// 5. 如果有背景图片（Base64 格式），上传到 PDS
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

	// 6. 更新 PDS 上的个人资料
	putRecordParams := map[string]interface{}{
		"repo":       aster.Did,
		"collection": "app.vtri.aster.profile",
		"rkey":       "self",
		"record":     profile,
	}

	var putResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, "application/json", "com.atproto.repo.putRecord",
		nil, putRecordParams, &putResult)

	if err != nil {
		return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "更新 PDS 个人资料失败: " + err.Error()})
	}

	// 7. 更新本地数据库
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := a.metaStore.DB.Model(&database.Avatar{}).Where("did = ?", aster.Did).Updates(updates).Error; err != nil {
			// 数据库更新失败，但 PDS 更新已成功
			return ac.JSON(http.StatusOK, map[string]interface{}{
				"success": true,
				"warning": "PDS 个人资料已更新，但本地数据库更新失败",
				"pds":     putResult,
			})
		}

		// 更新内存中的aster对象
		if displayName, ok := updates["display_name"].(string); ok && displayName != "" {
			aster.DisplayName = displayName
		}

		if description, ok := updates["description"].(string); ok && description != "" {
			aster.Description = description
		}

		if avatarCID, ok := updates["avatar_cid"].(string); ok && avatarCID != "" {
			aster.AvatarCID = avatarCID
		}
	}

	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.CreatorDid,
			aster.AvatarCID)
	}

	response := AsterProfileResponse{
		Did:         aster.Did,
		Handle:      aster.Handle,
		DisplayName: aster.DisplayName,
		Description: aster.Description,
		Avatar:      avatarURL,
		CreatedAt:   aster.CreatedAt.Format(time.RFC3339),
	}

	return ac.JSON(http.StatusOK, response)
}

func (a *AvatarAIAPI) HandleAsterMint(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	oauthSession := ac.OauthSession

	existingAster, err := database.GetAsterByCreatorDid(a.metaStore.DB, avatar.Did)
	if err == nil && existingAster != nil {
		return c.JSON(400, map[string]string{
			"error": "已经拥有Aster，不能重复铸造",
		})
	} else if err != nil && !errors.Is(err, database.ErrAsterNotFound) {
		return c.JSON(500, map[string]string{
			"error": "检查Aster状态失败: " + err.Error(),
		})
	}

	did, _, err := utils.GenerateDIDKey()
	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "生成 didKey 失败: " + err.Error(),
		})
	}

	imageData, err := mint.MintNFT(c.Request().Context(), did)
	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "生成个性化的图片失败: " + err.Error(),
		})
	}

	// 创建 xrpc 客户端
	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	// 上传生成的图片到 PDS
	mimeType := http.DetectContentType(imageData)
	var uploadResult struct {
		Blob *util.LexBlob `json:"blob"`
	}

	err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
		nil, bytes.NewReader(imageData), &uploadResult)

	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "上传Aster头像到PDS失败: " + err.Error(),
		})
	}

	displayName := fmt.Sprintf("%s's Aster", avatar.DisplayName)
	description := fmt.Sprintf("This is %s's Aster", avatar.DisplayName)

	// 创建 Aster Profile 记录
	profile := &vtri.AsterProfile{
		LexiconTypeID: "app.vtri.aster.profile",
		Did:           &did,
		Handle:        &did, // 临时使用 did 作为 handle
		Avatar:        uploadResult.Blob,
		DisplayName:   &displayName,
		Description:   &description,
	}

	// 设置创建时间
	createdAt := time.Now().Format(time.RFC3339)
	profile.CreatedAt = &createdAt

	// 设置creator引用
	creatorUri := fmt.Sprintf("at://%s/app.vtri.avatar.profile/self", avatar.Did)
	profile.Creator = &comatproto.RepoStrongRef{
		LexiconTypeID: "com.atproto.repo.strongRef",
		Uri:           creatorUri,
	}

	// 写入 PDS 记录
	putRecordParams := map[string]interface{}{
		"repo":       avatar.Did,
		"collection": "app.vtri.aster.profile",
		"rkey":       did,
		"record":     profile,
	}

	var putResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, "application/json", "com.atproto.repo.putRecord",
		nil, putRecordParams, &putResult)

	if err != nil {
		return c.JSON(500, map[string]string{
			"error": "创建Aster PDS记录失败: " + err.Error(),
		})
	}

	// 更新数据库中的 Aster 记录
	aster := &database.Avatar{
		Did:         did,
		CreatorDid:  avatar.Did,
		IsAster:     true,
		AvatarCID:   uploadResult.Blob.Ref.String(),
		DisplayName: displayName,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := database.CreateAster(a.metaStore.DB, aster); err != nil {
		return c.JSON(500, map[string]string{
			"error": "保存Aster失败: " + err.Error(),
		})
	}

	// 在成功创建数据库记录后，构建头像URL
	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.CreatorDid,
			aster.AvatarCID)
	}

	// 构建 AsterProfileResponse
	asterResponse := &AsterProfileResponse{
		Did:         aster.Did,
		Handle:      aster.Did, // 临时使用 did 作为 handle
		DisplayName: aster.DisplayName,
		Description: aster.Description,
		Avatar:      avatarURL,
		CreatedAt:   aster.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(200, GetAsterProfileResponse{
		Aster:       asterResponse,
		Initialized: true,
	})
}
