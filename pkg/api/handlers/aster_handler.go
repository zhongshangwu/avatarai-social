package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	indigo "github.com/bluesky-social/indigo/api/atproto"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/ipfs/go-cid"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/mint"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
	"github.com/zhongshangwu/avatarai-social/types"
)

type GetAsterProfileResponse struct {
	Aster       *UserProfileView `json:"aster"`
	Initialized bool             `json:"initialized"`
}

type UpdateAsterProfileRequest struct {
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	AvatarFileID string `json:"avatarFileId,omitempty"`
	BannerFileID string `json:"bannerFileId,omitempty"`
}

type UpdateAsterProfileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AsterHandler struct {
	config    *config.SocialConfig
	metaStore *repositories.MetaStore
}

func NewAsterHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *AsterHandler {
	return &AsterHandler{
		config:    config,
		metaStore: metaStore,
	}
}

func (h *AsterHandler) GetAsterProfile(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}
	user := c.User

	aster, err := h.metaStore.UserRepo.GetAsterByCreatorDid(user.Did)
	if err != nil {
		if errors.Is(err, repositories.ErrAsterNotFound) {
			return c.JSON(200, GetAsterProfileResponse{
				Aster:       nil,
				Initialized: false,
			})
		}
		return c.InternalServerError("获取Aster信息失败: " + err.Error())
	}

	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.Creator,
			aster.AvatarCID)
	}

	profile := &UserProfileView{
		Did:         aster.Did,
		Handle:      aster.Handle,
		DisplayName: aster.DisplayName,
		Description: aster.Description,
		Avatar:      avatarURL,
		AvatarCID:   aster.AvatarCID,
		CreatedAt:   aster.CreatedAt,
	}

	return c.JSON(200, GetAsterProfileResponse{
		Aster:       profile,
		Initialized: true,
	})
}

func (h *AsterHandler) HandleAsterUpdateProfile(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}

	var updateReq UpdateAsterProfileRequest
	if err := c.Bind(&updateReq); err != nil {
		return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "Invalid request data: "+err.Error())
	}

	user := c.User

	aster, err := h.metaStore.UserRepo.GetAsterByCreatorDid(user.Did)
	if err != nil {
		if errors.Is(err, repositories.ErrAsterNotFound) {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "您还没有创建Aster")
		}
		return c.InternalServerError("获取Aster信息失败: " + err.Error())
	}

	xrpcCli, err := atproto.NewXrpcClient(c.OauthSession)
	if err != nil {
		return c.InternalServerError("创建 XRPC 客户端失败: " + err.Error())
	}

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

	creatorUri := fmt.Sprintf("at://%s/app.vtri.avatar.profile/self", user.Did)
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

	// 处理头像上传 - 优先使用文件ID，其次使用base64
	if updateReq.AvatarFileID != "" {
		file, err := h.metaStore.FileRepo.GetUploadFileByID(updateReq.AvatarFileID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的头像文件ID: "+err.Error())
		}
		refCid, err := cid.Decode(file.CID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的文件CID: "+err.Error())
		}

		avatarBlob := &lexutil.LexBlob{
			Ref:      lexutil.LexLink(refCid),
			MimeType: file.MimeType,
			Size:     file.Size,
		}
		profile.Avatar = avatarBlob
		updates["avatar_cid"] = file.CID
	}

	// 处理背景图上传 - 优先使用文件ID，其次使用base64
	if updateReq.BannerFileID != "" {
		file, err := h.metaStore.FileRepo.GetUploadFileByID(updateReq.BannerFileID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的背景图文件ID: "+err.Error())
		}

		refCid, err := cid.Decode(file.CID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的文件CID: "+err.Error())
		}

		bannerBlob := &lexutil.LexBlob{
			Ref:      lexutil.LexLink(refCid),
			MimeType: file.MimeType,
			Size:     file.Size,
		}
		profile.Banner = bannerBlob
		updates["banner_cid"] = file.CID
	}

	rec := &lexutil.LexiconTypeDecoder{Val: profile}
	input := indigo.RepoPutRecord_Input{
		Collection: "app.vtri.aster.profile",
		Rkey:       "self",
		Repo:       aster.Did,
		Record:     rec,
	}
	output := indigo.RepoPutRecord_Output{}
	err = xrpcCli.Procedure(c.Request().Context(), "com.atproto.repo.putRecord", nil, input, &output)
	if err != nil {
		return c.InternalServerError("更新 PDS 个人资料失败: " + err.Error())
	}

	if len(updates) > 0 {
		if err := h.metaStore.UserRepo.UpdateAvatar(aster.Did, updates); err != nil {
			return c.JSON(http.StatusOK, UpdateAsterProfileResponse{
				Success: true,
				Message: "PDS 个人资料已更新，但本地数据库更新失败: " + err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, UpdateAsterProfileResponse{
		Success: true,
		Message: "Aster 个人资料更新成功",
	})
}

func (h *AsterHandler) HandleAsterMint(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}

	user := c.User

	existingAster, err := h.metaStore.UserRepo.GetAsterByCreatorDid(user.Did)
	if err == nil && existingAster != nil {
		return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "已经拥有Aster，不能重复铸造")
	} else if err != nil && !errors.Is(err, repositories.ErrAsterNotFound) {
		return c.InternalServerError("检查Aster状态失败: " + err.Error())
	}

	did, _, err := utils.GenerateDIDKey()
	if err != nil {
		return c.InternalServerError("生成 didKey 失败: " + err.Error())
	}

	imageData, err := mint.MintNFT(c.Request().Context(), did)
	if err != nil {
		return c.InternalServerError("生成个性化的图片失败: " + err.Error())
	}

	xrpcCli, err := atproto.NewXrpcClient(c.OauthSession)
	if err != nil {
		return c.InternalServerError("创建 XRPC 客户端失败: " + err.Error())
	}

	mimeType := http.DetectContentType(imageData)

	// 上传到 PDS
	var uploadResult indigo.RepoUploadBlob_Output

	err = xrpcCli.ProcedureWithEncoding(c.Request().Context(), "com.atproto.repo.uploadBlob", mimeType,
		nil, bytes.NewReader(imageData), &uploadResult)
	if err != nil {
		return c.InternalServerError("上传Aster头像到PDS失败: " + err.Error())
	}

	displayName := fmt.Sprintf("%s's Aster", user.DisplayName)
	description := fmt.Sprintf("This is %s's Aster", user.DisplayName)

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

	creatorUri := fmt.Sprintf("at://%s/app.vtri.avatar.profile/self", user.Did)
	profile.Creator = &comatproto.RepoStrongRef{
		LexiconTypeID: "com.atproto.repo.strongRef",
		Uri:           creatorUri,
	}

	rec := &lexutil.LexiconTypeDecoder{Val: profile}
	putInput := indigo.RepoPutRecord_Input{
		Collection: "app.vtri.aster.profile",
		Rkey:       did,
		Repo:       user.Did,
		Record:     rec,
	}
	putOutput := indigo.RepoPutRecord_Output{}

	err = xrpcCli.Procedure(c.Request().Context(), "com.atproto.repo.putRecord", nil, putInput, &putOutput)
	if err != nil {
		return c.InternalServerError("创建Aster PDS记录失败: " + err.Error())
	}

	aster := &repositories.Avatar{
		Did:         did,
		Creator:     user.Did,
		IsAster:     true,
		AvatarCID:   uploadResult.Blob.Ref.String(),
		DisplayName: displayName,
		Description: description,
		CreatedAt:   utils.Timestamp(),
	}

	if err := h.metaStore.UserRepo.CreateAster(aster); err != nil {
		return c.InternalServerError("保存Aster失败: " + err.Error())
	}

	avatarURL := ""
	if aster.AvatarCID != "" {
		avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
			aster.Creator,
			aster.AvatarCID)
	}

	asterResponse := &UserProfileView{
		Did:         aster.Did,
		Handle:      aster.Did,
		DisplayName: aster.DisplayName,
		Description: aster.Description,
		Avatar:      avatarURL,
		AvatarCID:   aster.AvatarCID,
		CreatedAt:   aster.CreatedAt,
	}

	return c.JSON(200, GetAsterProfileResponse{
		Aster:       asterResponse,
		Initialized: true,
	})
}
