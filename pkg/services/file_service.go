package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	indigo "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/lex/util"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/gabriel-vasile/mimetype"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/blobs"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

const (
	MaxFileSize = 50 * 1024 * 1024 // 50MB
	UnamedFile  = "unamed_file"
)

var SupportedMimeTypes = map[string]bool{
	"image/jpeg":       true,
	"image/png":        true,
	"image/gif":        true,
	"image/webp":       true,
	"video/mp4":        true,
	"video/webm":       true,
	"video/quicktime":  true,
	"audio/mpeg":       true,
	"audio/wav":        true,
	"application/pdf":  true,
	"text/plain":       true,
	"application/json": true,
	"application/xml":  true,
	"application/zip":  true,
	"application/rar":  true,
	"application/7z":   true,
	"application/tar":  true,
	"application/gz":   true,
}

type FileService struct {
	metaStore    *repositories.MetaStore
	imageBuilder *blobs.ImageUriBuilder
}

type UploadFileResponse struct {
	ID        string       `json:"id"`
	Filename  string       `json:"filename"`
	Extension string       `json:"extension"`
	MimeType  string       `json:"mime_type"`
	Size      int64        `json:"size"`
	CID       string       `json:"cid"`
	URL       string       `json:"url"`
	CreatedBy string       `json:"created_by"`
	CreatedAt time.Time    `json:"created_at"`
	Blob      util.LexBlob `json:"blob"`
}

type FileListResponse struct {
	Files      []*UploadFileResponse `json:"files"`
	Pagination *PaginationInfo       `json:"pagination"`
}

type PaginationInfo struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

func NewFileService(config *config.SocialConfig, metaStore *repositories.MetaStore) *FileService {
	return &FileService{
		metaStore:    metaStore,
		imageBuilder: blobs.NewImageUriBuilder(config.Server.Domain),
	}
}

func (s *FileService) UploadFile(
	ctx context.Context,
	userDid string,
	oauthSession *types.OAuthSession,
	content io.Reader,
	filename string) (*types.UploadFile, error) {
	if oauthSession == nil {
		return nil, fmt.Errorf("用户没有有效的OAuth会话")
	}

	fileBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	logrus.Infof("upload fileBytes: %d", len(fileBytes))

	if s.IsFileSizeLimited(len(fileBytes)) {
		return nil, fmt.Errorf("文件大小超过限制")
	}

	if filename == "" {
		filename = UnamedFile
	}

	mimeType, extension := s.detectMimeTypeAndExtension(filename, fileBytes)
	if !s.IsValidFileType(mimeType) {
		return nil, fmt.Errorf("不支持的文件类型: %s", mimeType)
	}

	xrpcCli, err := atproto.NewXrpcClient(oauthSession, atproto.WithNonceUpdateCallback(func(did, newNonce string) error {
		return s.metaStore.OAuthRepo.UpdateOAuthSessionDpopPdsNonce(did, newNonce)
	}))
	if err != nil {
		return nil, fmt.Errorf("创建 XRPC 客户端失败: %w", err)
	}

	var uploadResult indigo.RepoUploadBlob_Output
	err = xrpcCli.ProcedureWithEncoding(ctx, "com.atproto.repo.uploadBlob", mimeType,
		nil, fileBytes, &uploadResult)
	if err != nil {
		return nil, fmt.Errorf("上传文件到 PDS 失败: %w", err)
	}

	logrus.Infof("uploadBlob: %+v, mimeType: %s", uploadResult.Blob, mimeType)

	if uploadResult.Blob.Ref.String() == "" {
		return nil, fmt.Errorf("PDS 上传结果为空")
	}

	cid := uploadResult.Blob.Ref.String()

	fileID := s.GenerateFileID()
	now := time.Now().Unix()
	rec := &lexutil.LexiconTypeDecoder{Val: &vtri.EntityFile{
		LexiconTypeID: "app.vtri.entity.file",
		Blob:          uploadResult.Blob,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}}
	putInput := indigo.RepoPutRecord_Input{
		Collection: "app.vtri.entity.file",
		Rkey:       fileID,
		Repo:       userDid,
		Record:     rec,
	}
	putOutput := indigo.RepoPutRecord_Output{}

	err = xrpcCli.Procedure(ctx, "com.atproto.repo.putRecord", nil, putInput, &putOutput)
	if err != nil {
		return nil, fmt.Errorf("上传文件记录失败: %w", err)
	}

	uploadFile := &repositories.UploadFile{
		ID:        fileID,
		CID:       putOutput.Cid,
		URI:       putOutput.Uri,
		BlobCID:   uploadResult.Blob.Ref.String(),
		Size:      int64(len(fileBytes)),
		Filename:  filename,
		Extension: extension,
		MimeType:  mimeType,
		CreatedBy: userDid,
		CreatedAt: now,
	}

	if err := s.metaStore.FileRepo.CreateUploadFile(uploadFile); err != nil {
		return nil, fmt.Errorf("保存文件记录失败: %w", err)
	}

	url, err := s.imageBuilder.GetPresetUri(blobs.PresetAvatar, userDid, cid)
	if err != nil {
		return nil, fmt.Errorf("获取文件URL失败: %w", err)
	}

	return &types.UploadFile{
		ID:        fileID,
		Filename:  filename,
		Extension: extension,
		MimeType:  mimeType,
		Size:      int64(len(fileBytes)),
		CID:       cid,
		URL:       url,
		CreatedBy: userDid,
		CreatedAt: uploadFile.CreatedAt,
	}, nil
}

func (s *FileService) GetFile(ctx context.Context, fileCid string) (*types.UploadFile, error) {
	uploadFile, err := s.metaStore.FileRepo.GetUploadFileByBlobCID(fileCid)
	if err != nil {
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}

	oauthSession, err := s.metaStore.OAuthRepo.GetOAuthSessionByDID(uploadFile.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("获取OAuth会话失败: %w", err)
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s",
		oauthSession.PdsUrl, oauthSession.Did, uploadFile.CID)

	return &types.UploadFile{
		ID:        uploadFile.ID,
		Filename:  uploadFile.Filename,
		Extension: uploadFile.Extension,
		MimeType:  uploadFile.MimeType,
		Size:      uploadFile.Size,
		CID:       uploadFile.CID,
		URL:       url,
		CreatedBy: uploadFile.CreatedBy,
		CreatedAt: uploadFile.CreatedAt,
	}, nil
}

func (s *FileService) IsValidFileType(mimeType string) bool {
	return SupportedMimeTypes[mimeType]
}

func (s *FileService) IsFileSizeLimited(size int) bool {
	return size > MaxFileSize
}

func (s *FileService) GenerateFileID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *FileService) detectMimeTypeAndExtension(filename string, fileBytes []byte) (string, string) {
	mtype := mimetype.Detect(fileBytes)
	detectedMimeType := mtype.String()
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = mtype.Extension()
	}
	if ext != "" && ext[0] == '.' {
		ext = ext[1:]
	}
	return detectedMimeType, ext
}
