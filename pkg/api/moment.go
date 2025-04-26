package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	comatprototypes "github.com/bluesky-social/indigo/api/atproto"
	appbskytypes "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type CreateMomentRequest struct {
	Text     string                         `json:"text"`               // 文本内容
	Facets   []*appbskytypes.RichtextFacet  `json:"facets,omitempty"`   // 富文本注解
	Root     *comatprototypes.RepoStrongRef `json:"root,omitempty"`     // 根帖子引用
	Parent   *comatprototypes.RepoStrongRef `json:"parent,omitempty"`   // 父帖子引用
	Images   []util.LexBlob                 `json:"images,omitempty"`   // 图片引用
	Video    *util.LexBlob                  `json:"video,omitempty"`    // 视频引用
	External *vtri.EntityExternal_External  `json:"external,omitempty"` // 外部链接
	Langs    []string                       `json:"langs,omitempty"`    // 语言标签
	Tags     []string                       `json:"tags,omitempty"`     // 标签
}

func (a *AvatarAIAPI) HandleMomentCreate(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession

	var req CreateMomentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "请求格式错误："+err.Error())
	}

	log.Println("request json", req)

	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	moment := &vtri.ActivityMoment{
		LexiconTypeID: "app.vtri.activity.moment",
		Text:          req.Text,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Langs:         req.Langs,
		Tags:          req.Tags,
	}

	if len(req.Facets) > 0 {
		moment.Facets = req.Facets

		for _, facet := range req.Facets {
			log.Println(fmt.Sprintf("facet: %+v", facet.Features[0]))
		}
	}

	if req.Root != nil {
		moment.Reply = &vtri.ActivityMoment_ReplyRef{
			Root:   req.Root,
			Parent: req.Parent,
		}
	}

	if len(req.Images) > 0 {
		images := &vtri.EntityImages{
			LexiconTypeID: "app.vtri.entity.images",
			Images:        make([]*vtri.EntityImages_Image, 0, len(req.Images)),
		}

		for _, img := range req.Images {
			images.Images = append(images.Images, &vtri.EntityImages_Image{
				Image: &img,
				Alt:   "", // 假设 LexBlob 有 Alt 字段，如果没有可能需要额外处理
			})
		}

		moment.Embed = &vtri.ActivityMoment_Embed{
			EntityImages: images,
		}
	} else if req.Video != nil {
		video := &vtri.EntityVideo{
			LexiconTypeID: "app.vtri.entity.video",
			Video:         req.Video,
			Alt:           nil, // 同上，根据实际结构调整
		}

		moment.Embed = &vtri.ActivityMoment_Embed{
			EntityVideo: video,
		}
	} else if req.External != nil {
		external := &vtri.EntityExternal{
			LexiconTypeID: "app.vtri.entity.external",
			External:      req.External,
		}

		moment.Embed = &vtri.ActivityMoment_Embed{
			EntityExternal: external,
		}
	}

	rkey := helper.GenerateTID()

	putRecordParams := map[string]interface{}{
		"repo":       oauthSession.Did,
		"collection": "app.vtri.activity.moment",
		"rkey":       rkey,
		"record":     moment,
	}

	var putResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	err := xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Procedure, "application/json", "com.atproto.repo.putRecord",
		nil, putRecordParams, &putResult)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "创建 ATProto 记录失败: "+err.Error())
	}

	atUri := putResult.URI
	atCid := putResult.CID

	tx := a.metaStore.DB.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "开始数据库事务失败: "+tx.Error.Error())
	}

	momentJSON, err := json.Marshal(moment)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "序列化 moment 失败: "+err.Error())
	}

	atpRecord := &database.AtpRecord{
		URI:       atUri,
		CID:       atCid,
		Did:       oauthSession.Did,
		JSON:      string(momentJSON),
		IndexedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := database.InsertOrUpdateAtpRecord(tx, atpRecord); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "保存 ATProto 记录失败: "+err.Error())
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	dbMoment := &database.Moment{
		URI:       atUri,
		CID:       atCid,
		Creator:   oauthSession.Did,
		Text:      req.Text,
		Langs:     req.Langs,
		Tags:      req.Tags,
		CreatedAt: nowStr,
		IndexedAt: nowStr,
		SortAt:    nowStr,
	}

	if req.Root != nil {
		dbMoment.ReplyRoot = req.Root.Uri
		dbMoment.ReplyRootCID = req.Root.Cid
		dbMoment.ReplyParent = req.Parent.Uri
		dbMoment.ReplyParentCID = req.Parent.Cid
	}

	if err := database.CreateMoment(tx, dbMoment); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "保存 Moment 记录失败: "+err.Error())
	}

	if len(req.Images) > 0 {
		for i, img := range req.Images {
			dbImage := &database.MomentImage{
				MomentURI: atUri,
				Position:  i,
				ImageCID:  img.Ref.String(),
				Alt:       "",
			}
			if err := database.CreateMomentImage(tx, dbImage); err != nil {
				tx.Rollback()
				return echo.NewHTTPError(http.StatusInternalServerError, "保存图片记录失败: "+err.Error())
			}
		}
	}

	if req.Video != nil {
		dbVideo := &database.MomentVideo{
			MomentURI: atUri,
			VideoCID:  req.Video.Ref.String(),
			Alt:       "",
		}
		if err := database.CreateMomentVideo(tx, dbVideo); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "保存视频记录失败: "+err.Error())
		}
	}

	if req.External != nil {
		dbExternal := &database.MomentExternal{
			MomentURI:   atUri,
			URI:         req.External.Uri,
			Title:       req.External.Title,
			Description: req.External.Description,
		}
		if req.External.Thumb != nil {
			dbExternal.ThumbCID = req.External.Thumb.Ref.String()
		}
		if err := database.CreateMomentExternal(tx, dbExternal); err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "保存外部链接记录失败: "+err.Error())
		}
	}

	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "提交事务失败: "+err.Error())
	}

	uris := []string{atUri}
	hydrationState, err := a.hydrateMoments(c.Request().Context(), uris, xrpcCli, authArgs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "水合 moments 失败: "+err.Error())
	}

	feedItems := a.presentMoments(uris, hydrationState)
	return c.JSON(http.StatusCreated, feedItems[0].Moment)
}

type FeedItem struct {
	Type   string      `json:"type"`
	Moment *MomentView `json:"moment"`
}

type FeedResponse struct {
	Cursor string      `json:"cursor"`
	Feed   []*FeedItem `json:"feed"`
}

type MomentView struct {
	URI       string        `json:"uri"`
	CID       string        `json:"cid"`
	Author    *AuthorView   `json:"author"`
	Record    *MomentRecord `json:"record"`
	Embed     *EmbedView    `json:"embed,omitempty"`
	IndexedAt time.Time     `json:"indexedAt"`
	Labels    []string      `json:"labels,omitempty"`
}

type MomentRecord struct {
	Type      string                        `json:"$type"`
	Text      string                        `json:"text"`
	CreatedAt string                        `json:"createdAt"`
	Reply     *vtri.ActivityMoment_ReplyRef `json:"reply,omitempty"`
	Embed     *vtri.ActivityMoment_Embed    `json:"embed,omitempty"`
	Facets    []*appbskytypes.RichtextFacet `json:"facets,omitempty"`
	Langs     []string                      `json:"langs,omitempty"`
	Tags      []string                      `json:"tags,omitempty"`
}

type AuthorView struct {
	DID         string    `json:"did"`
	Handle      string    `json:"handle"`
	DisplayName string    `json:"displayName"`
	Avatar      string    `json:"avatar,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type EmbedView struct {
	Type     string        `json:"$type"`
	Images   []*ImageView  `json:"images,omitempty"`
	Video    *VideoView    `json:"video,omitempty"`
	External *ExternalView `json:"external,omitempty"`
	Record   *RecordView   `json:"record,omitempty"`
}

type ImageView struct {
	Thumb    string `json:"thumb"`
	Fullsize string `json:"fullsize"`
	Alt      string `json:"alt,omitempty"`
}

type VideoView struct {
	Thumb string `json:"thumb,omitempty"`
	Video string `json:"video"`
	Alt   string `json:"alt,omitempty"`
}

type ExternalView struct {
	URI         string `json:"uri"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Thumb       string `json:"thumb,omitempty"`
}

type RecordView struct {
	URI    string      `json:"uri"`
	CID    string      `json:"cid"`
	Author *AuthorView `json:"author"`
	Value  interface{} `json:"value"`
}

type MomentFeedResponse struct {
	Feed   []*FeedItem `json:"feed"`
	Cursor string      `json:"cursor,omitempty"`
}

type FeedParams struct {
	Limit  int    `query:"limit"`
	Cursor string `query:"cursor"`
	Feed   string `query:"feed"`
}

func (a *AvatarAIAPI) HandleMomentDetail(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession

	uri := c.QueryParam("uri")
	if uri == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "无效的请求参数: uri")
	}

	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)

	_, err := database.GetAtpRecord(a.metaStore.DB, uri)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取 ATProto 记录失败: "+err.Error())
	}

	hydrationState, err := a.hydrateMoments(c.Request().Context(), []string{uri}, xrpcCli, authArgs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "水合 moments 失败: "+err.Error())
	}

	feedItems := a.presentMoments([]string{uri}, hydrationState)
	return c.JSON(http.StatusCreated, feedItems[0].Moment)
}

func (a *AvatarAIAPI) HandleMomentFeed(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession

	params := new(FeedParams)
	if err := c.Bind(params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "无效的请求参数: "+err.Error())
	}

	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 30 // 默认限制
	}

	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)
	viewerDID := oauthSession.Did

	uris, nextCursor, err := a.getFeedURIs(c.Request().Context(), params, xrpcCli, authArgs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取 feed URIs 失败: "+err.Error())
	}

	hydrationState, err := a.hydrateMoments(c.Request().Context(), uris, xrpcCli, authArgs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "水合 moments 失败: "+err.Error())
	}

	filteredURIs := a.filterBlockedContent(uris, hydrationState, viewerDID)

	feedItems := a.presentMoments(filteredURIs, hydrationState)

	response := &MomentFeedResponse{
		Feed:   feedItems,
		Cursor: nextCursor,
	}

	return c.JSON(http.StatusOK, response)
}

func (a *AvatarAIAPI) getFeedURIs(ctx context.Context, params *FeedParams, client *atproto.XrpcClient, authArgs *atproto.XrpcAuthedRequestArgs) ([]string, string, error) {
	var uris []string
	var nextCursor string

	if strings.HasPrefix(params.Feed, "at://") {
		uris = []string{
			"at://did:plc:example1/app.vtri.activity.moment/1",
			"at://did:plc:example2/app.vtri.activity.moment/2",
			"at://did:plc:example3/app.vtri.activity.moment/3",
		}
		nextCursor = "mock_cursor_value"
	} else {
		moments, err := database.GetLatestMoments(a.metaStore.DB, params.Limit, params.Cursor)
		if err != nil {
			return nil, "", err
		}

		for _, moment := range moments {
			uris = append(uris, moment.URI)
		}

		if len(moments) >= params.Limit {
			nextCursor = generateCursor(moments[len(moments)-1])
		}
	}

	return uris, nextCursor, nil
}

func generateCursor(moment *database.Moment) string {
	cursorData := map[string]interface{}{
		"t": moment.SortAt,
		"u": moment.URI,
	}
	jsonBytes, _ := json.Marshal(cursorData)
	return base64.StdEncoding.EncodeToString(jsonBytes)
}

func (a *AvatarAIAPI) hydrateMoments(ctx context.Context, uris []string, client *atproto.XrpcClient, authArgs *atproto.XrpcAuthedRequestArgs) (map[string]interface{}, error) {
	hydrationState := make(map[string]interface{})

	batchSize := 25
	for i := 0; i < len(uris); i += batchSize {
		end := i + batchSize
		if end > len(uris) {
			end = len(uris)
		}
		batchURIs := uris[i:end]

		records, err := database.GetAtpRecords(a.metaStore.DB, batchURIs)
		if err != nil {
			return nil, err
		}

		for _, record := range records {

			aturi, err := helper.BuildAtURI(record.URI)
			if err != nil {
				log.Printf("无法解析记录 URI: %s", err)
				continue // 跳过无法解析的记录
			}

			if aturi.Collection() != "app.vtri.activity.moment" {
				log.Printf("记录 URI 不是 moment: %s", record.URI)
				continue // 跳过非 moment 记录
			}

			var momentRecord vtri.ActivityMoment
			if err := json.Unmarshal([]byte(record.JSON), &momentRecord); err != nil {
				continue // 跳过无法解析的记录
			}

			hydrationState[record.URI] = map[string]interface{}{
				"uri":    record.URI,
				"cid":    record.CID,
				"author": record.Did,
				"value":  momentRecord,
			}

			// 如果有嵌入内容或引用，可能需要额外获取这些内容
			// 例如：获取引用的记录、图片元数据等
		}
	}

	// 获取关联的用户资料信息
	if err := a.hydrateProfiles(ctx, hydrationState, client, authArgs); err != nil {
		return nil, err
	}

	return hydrationState, nil
}

func (a *AvatarAIAPI) hydrateProfiles(ctx context.Context, hydrationState map[string]interface{}, client *atproto.XrpcClient, authArgs *atproto.XrpcAuthedRequestArgs) error {
	dids := make(map[string]bool)
	for _, data := range hydrationState {
		if recordData, ok := data.(map[string]interface{}); ok {
			if author, ok := recordData["author"].(string); ok {
				dids[author] = true
			}
		}
	}

	if len(dids) == 0 {
		return nil
	}

	didsList := make([]string, 0, len(dids))
	for did := range dids {
		didsList = append(didsList, did)
	}

	batchSize := 25
	for i := 0; i < len(didsList); i += batchSize {
		end := i + batchSize
		if end > len(didsList) {
			end = len(didsList)
		}
		batchDIDs := didsList[i:end]

		avatars, err := database.GetAvatarsByDIDs(a.metaStore.DB, batchDIDs)
		if err != nil {
			log.Printf("获取用户资料失败: %s", err)
			return err
		}

		for _, avatar := range avatars {
			hydrationState["profile:"+avatar.Did] = avatar
		}
	}

	return nil
}

func (a *AvatarAIAPI) filterBlockedContent(uris []string, hydrationState map[string]interface{}, viewerDID string) []string {
	blockedDIDs, err := database.GetBlockedDIDs(a.metaStore.DB, viewerDID)
	if err != nil {
		log.Printf("获取屏蔽列表失败: %v", err)
		return uris
	}

	blockedMap := make(map[string]bool)
	for _, did := range blockedDIDs {
		blockedMap[did] = true
	}

	var filteredURIs []string
	for _, uri := range uris {
		if data, ok := hydrationState[uri].(map[string]interface{}); ok {
			if author, ok := data["author"].(string); ok {
				if !blockedMap[author] {
					filteredURIs = append(filteredURIs, uri)
				}
			}
		}
	}

	return filteredURIs
}

func (a *AvatarAIAPI) presentMoments(uris []string, hydrationState map[string]interface{}) []*FeedItem {
	var feedItems []*FeedItem

	for _, uri := range uris {
		data, ok := hydrationState[uri].(map[string]interface{})
		if !ok {
			continue
		}

		authorDID, _ := data["author"].(string)
		authorProfile, hasProfile := hydrationState["profile:"+authorDID].(*database.Avatar)

		authorView := &AuthorView{
			DID: authorDID,
		}

		if hasProfile {
			authorView.Handle = authorProfile.Handle
			authorView.DisplayName = authorProfile.DisplayName
			avatarURL := fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
				authorProfile.Did,
				authorProfile.AvatarCID)
			authorView.Avatar = avatarURL
			authorView.CreatedAt = authorProfile.CreatedAt
		}

		recordValue, _ := data["value"].(vtri.ActivityMoment)

		recordView := &MomentRecord{
			Type:      "app.vtri.activity.moment",
			Text:      recordValue.Text,
			CreatedAt: recordValue.CreatedAt,
			Facets:    recordValue.Facets,
			Langs:     recordValue.Langs,
			Tags:      recordValue.Tags,
			Reply:     recordValue.Reply,
		}

		var embedView *EmbedView
		if recordValue.Embed != nil {
			embedView = &EmbedView{}

			if recordValue.Embed.EntityImages != nil {
				embedView.Type = "app.vtri.entity.images#view"
				embedView.Images = make([]*ImageView, 0, len(recordValue.Embed.EntityImages.Images))

				for _, img := range recordValue.Embed.EntityImages.Images {
					imageView := &ImageView{
						Fullsize: fmt.Sprintf("https://bsky.avatar.ai/img/fullsize/%s", img.Image.Ref.String()),
						Thumb:    fmt.Sprintf("https://bsky.avatar.ai/img/thumb/%s", img.Image.Ref.String()),
						Alt:      img.Alt,
					}
					embedView.Images = append(embedView.Images, imageView)
				}
			} else if recordValue.Embed.EntityVideo != nil {
				embedView.Type = "app.vtri.entity.video#view"
				embedView.Video = &VideoView{
					Video: fmt.Sprintf("https://bsky.avatar.ai/video/%s", recordValue.Embed.EntityVideo.Video.Ref.String()),
					Alt:   "", // 需要设置 Alt 文本
				}
			} else if recordValue.Embed.EntityExternal != nil {
				embedView.Type = "app.vtri.entity.external#view"
				embedView.External = &ExternalView{
					URI:         recordValue.Embed.EntityExternal.External.Uri,
					Title:       recordValue.Embed.EntityExternal.External.Title,
					Description: recordValue.Embed.EntityExternal.External.Description,
				}

				if recordValue.Embed.EntityExternal.External.Thumb != nil {
					embedView.External.Thumb = fmt.Sprintf("https://bsky.avatar.ai/img/thumb/%s", recordValue.Embed.EntityExternal.External.Thumb.Ref.String())
				}
			}
		}

		indexedAt, _ := time.Parse(time.RFC3339, recordValue.CreatedAt)

		momentView := &MomentView{
			URI:       uri,
			CID:       data["cid"].(string),
			Author:    authorView,
			Record:    recordView,
			Embed:     embedView,
			IndexedAt: indexedAt,
		}

		feedItems = append(feedItems, &FeedItem{
			Type:   "app.vtri.activity.moment#view",
			Moment: momentView,
		})
	}

	return feedItems
}
