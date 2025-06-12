package handlers

import (
	"net/http"
	"time"

	indigo "github.com/bluesky-social/indigo/api/atproto"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/ipfs/go-cid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/blobs"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type SimpleUserView struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

type UserProfileView struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Banner      string `json:"banner,omitempty"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	AvatarCID   string `json:"avatarCID,omitempty"`
	BannerCID   string `json:"bannerCID,omitempty"`
}

type UpdateProfileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UserHandler struct {
	config       *config.SocialConfig
	metaStore    *repositories.MetaStore
	imageBuilder *blobs.ImageUriBuilder
}

func NewUserHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *UserHandler {
	return &UserHandler{
		config:       config,
		metaStore:    metaStore,
		imageBuilder: blobs.NewImageUriBuilder(config.Server.Domain),
	}
}

func (h *UserHandler) CurrentUserProfile(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}
	user := c.User

	avatarURL, _ := h.imageBuilder.GetPresetUri(blobs.PresetAvatar, user.Did, user.AvatarCID)
	bannerURL, _ := h.imageBuilder.GetPresetUri(blobs.PresetBanner, user.Did, user.BannerCID)
	response := UserProfileView{
		Did:         user.Did,
		Handle:      user.Handle,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Avatar:      avatarURL,
		Banner:      bannerURL,
		CreatedAt:   user.CreatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *UserHandler) UpdateUserProfile(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}

	var updateReq UpdateProfileRequest
	if err := c.Bind(&updateReq); err != nil {
		return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "Invalid request data: "+err.Error())
	}

	user := c.User

	xrpcCli, err := atproto.NewXrpcClient(c.OauthSession)
	if err != nil {
		return c.InternalServerError("创建 XRPC 客户端失败: " + err.Error())
	}

	profile := &vtri.AvatarProfile{
		LexiconTypeID: "app.vtri.avatar.profile",
	}

	did := user.Did
	handle := user.Handle

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

	// 为数据库准备更新
	updates := map[string]interface{}{}
	if updateReq.DisplayName != "" {
		updates["display_name"] = updateReq.DisplayName
	}

	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}

	if updateReq.AvatarCID != "" {
		file, err := h.metaStore.FileRepo.GetUploadFileByBlobCID(updateReq.AvatarCID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的头像文件ID: "+err.Error())
		}
		refCid, err := cid.Decode(file.BlobCID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的文件CID: "+err.Error())
		}

		avatarBlob := &lexutil.LexBlob{
			Ref:      lexutil.LexLink(refCid),
			MimeType: file.MimeType,
			Size:     file.Size,
		}
		profile.Avatar = avatarBlob
		updates["avatar_cid"] = file.BlobCID
	}

	// 处理背景图上传 - 优先使用文件ID，其次使用base64
	if updateReq.BannerCID != "" {
		// 通过文件ID获取文件信息
		file, err := h.metaStore.FileRepo.GetUploadFileByBlobCID(updateReq.BannerCID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的背景图文件ID: "+err.Error())
		}

		refCid, err := cid.Decode(file.BlobCID)
		if err != nil {
			return c.InvalidRequest(string(types.ErrorCodeInvalidRequestParams), "无效的文件CID: "+err.Error())
		}

		bannerBlob := &lexutil.LexBlob{
			Ref:      lexutil.LexLink(refCid),
			MimeType: file.MimeType,
			Size:     file.Size,
		}
		profile.Banner = bannerBlob
		updates["banner_cid"] = file.BlobCID
	}

	params := map[string]interface{}{
		"cid":        "",
		"collection": "app.vtri.avatar.profile",
		"repo":       user.Did,
		"rkey":       "self",
	}
	var ex indigo.RepoGetRecord_Output
	err = xrpcCli.Query(c.Request().Context(), "com.atproto.repo.getRecord", params, &ex)
	if err != nil {
		return c.InternalServerError("获取 PDS 个人资料失败: " + err.Error())
	}
	previousCid := ex.Cid
	logrus.Infof("previousCid: %s", *previousCid)
	rec := &lexutil.LexiconTypeDecoder{Val: profile}
	input := indigo.RepoPutRecord_Input{
		Collection: "app.vtri.avatar.profile",
		Rkey:       "self",
		Repo:       user.Did,
		Record:     rec,
		SwapRecord: previousCid,
	}
	output := indigo.RepoPutRecord_Output{}

	err = xrpcCli.Procedure(c.Request().Context(), "com.atproto.repo.putRecord", nil, input, &output)
	if err != nil {
		return c.InternalServerError("更新 PDS 个人资料失败: " + err.Error())
	}

	if len(updates) > 0 {
		if err := h.metaStore.UserRepo.UpdateAvatar(user.Did, updates); err != nil {
			return c.JSON(http.StatusOK, UpdateProfileResponse{
				Success: true,
				Message: "PDS 个人资料已更新，但本地数据库更新失败: " + err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, UpdateProfileResponse{
		Success: true,
		Message: "个人资料更新成功",
	})
}
