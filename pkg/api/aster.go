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

type GetAsterProfileResponse struct {
	Aster       *AvatarView `json:"aster"`
	Initialized bool        `json:"initialized"`
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

	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.CreatorDid,
			aster.AvatarCID)
	}

	profile := &AvatarView{
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

	aster, err := database.GetAsterByCreatorDid(a.metaStore.DB, avatar.Did)
	if err != nil {
		if errors.Is(err, database.ErrAsterNotFound) {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "您还没有创建Aster"})
		}
		return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "获取Aster信息失败: " + err.Error()})
	}

	var updateReq UpdateAsterProfileRequest
	if err := c.Bind(&updateReq); err != nil {
		return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的请求数据"})
	}

	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	profile := &vtri.AsterProfile{
		LexiconTypeID: "app.vtri.aster.profile",
	}

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

	createdAt := time.Now().Format(time.RFC3339)
	profile.CreatedAt = &createdAt

	creatorUri := fmt.Sprintf("at://%s/app.vtri.avatar.profile/self", avatar.Did)
	profile.Creator = &comatproto.RepoStrongRef{
		LexiconTypeID: "com.atproto.repo.strongRef",
		Uri:           creatorUri,
	}

	updates := map[string]interface{}{}
	if updateReq.DisplayName != "" {
		updates["display_name"] = updateReq.DisplayName
	}

	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}

	if updateReq.AvatarBase64 != "" {

		base64Data := updateReq.AvatarBase64
		if idx := strings.Index(base64Data, ","); idx != -1 {
			base64Data = base64Data[idx+1:]
		}

		imgData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的头像图片数据"})
		}

		mimeType := http.DetectContentType(imgData)
		if !strings.HasPrefix(mimeType, "image/") {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "上传的文件不是有效的图像"})
		}

		var uploadResult struct {
			Blob *util.LexBlob `json:"blob"`
		}

		err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
			nil, bytes.NewReader(imgData), &uploadResult)

		if err != nil {
			return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "上传头像到 PDS 失败: " + err.Error()})
		}

		if uploadResult.Blob != nil {
			profile.Avatar = uploadResult.Blob
			updates["avatar_cid"] = uploadResult.Blob.Ref.String()
		}
	}

	if updateReq.BannerBase64 != "" {

		base64Data := updateReq.BannerBase64
		if idx := strings.Index(base64Data, ","); idx != -1 {
			base64Data = base64Data[idx+1:]
		}

		imgData, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "无效的背景图片数据"})
		}

		mimeType := http.DetectContentType(imgData)
		if !strings.HasPrefix(mimeType, "image/") {
			return ac.JSON(http.StatusBadRequest, map[string]string{"error": "上传的文件不是有效的图像"})
		}

		var uploadResult struct {
			Blob *util.LexBlob `json:"blob"`
		}

		err = xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, mimeType, "com.atproto.repo.uploadBlob",
			nil, bytes.NewReader(imgData), &uploadResult)

		if err != nil {
			return ac.JSON(http.StatusInternalServerError, map[string]string{"error": "上传背景图到 PDS 失败: " + err.Error()})
		}

		if uploadResult.Blob != nil {
			profile.Banner = uploadResult.Blob
		}
	}

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

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := a.metaStore.DB.Model(&database.Avatar{}).Where("did = ?", aster.Did).Updates(updates).Error; err != nil {

			return ac.JSON(http.StatusOK, map[string]interface{}{
				"success": true,
				"warning": "PDS 个人资料已更新，但本地数据库更新失败",
				"pds":     putResult,
			})
		}

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

	response := AvatarView{
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

	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

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

	profile := &vtri.AsterProfile{
		LexiconTypeID: "app.vtri.aster.profile",
		Did:           &did,
		Handle:        &did,
		Avatar:        uploadResult.Blob,
		DisplayName:   &displayName,
		Description:   &description,
	}

	createdAt := time.Now().Format(time.RFC3339)
	profile.CreatedAt = &createdAt

	creatorUri := fmt.Sprintf("at://%s/app.vtri.avatar.profile/self", avatar.Did)
	profile.Creator = &comatproto.RepoStrongRef{
		LexiconTypeID: "com.atproto.repo.strongRef",
		Uri:           creatorUri,
	}

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

	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.CreatorDid,
			aster.AvatarCID)
	}

	asterResponse := &AvatarView{
		Did:         aster.Did,
		Handle:      aster.Did,
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
