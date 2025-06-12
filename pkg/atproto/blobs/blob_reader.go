package blobs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"golang.org/x/xerrors"
)

// BlobReader 通用的blob读取器接口
type BlobReader interface {
	// StreamBlob 流式读取blob数据
	StreamBlob(ctx context.Context, options BlobOptions, processor BlobProcessor) error

	// GetBlob 直接获取blob数据
	GetBlob(ctx context.Context, options BlobOptions) (*BlobData, error)

	// VerifyBlob 验证blob的完整性
	VerifyBlob(ctx context.Context, options BlobOptions) error
}

// BlobOptions 通用的blob选项
type BlobOptions struct {
	// 必填字段
	Identifier string // 可以是CID、文件路径、URL等

	// 可选字段
	DID            string            // ATProto DID
	CID            string            // IPFS CID
	StorageType    StorageType       // 存储类型
	AcceptEncoding string            // 接受的编码
	Context        context.Context   // 上下文
	Headers        map[string]string // 自定义头部

	// 验证选项
	SkipVerification bool  // 是否跳过验证
	ExpectedSize     int64 // 期望的文件大小
}

// BlobData 通用的blob数据
type BlobData struct {
	Data        []byte            // 原始数据
	ContentType string            // 内容类型
	Size        int64             // 数据大小
	Metadata    map[string]string // 元数据
	Source      string            // 数据来源
}

// BlobProcessor blob处理器函数类型
type BlobProcessor func(data *BlobData) error

// StorageType 存储类型枚举
type StorageType string

const (
	StorageATProto StorageType = "atproto" // ATProto数据平面
	StorageHTTP    StorageType = "http"    // HTTP URL
	StorageLocal   StorageType = "local"   // 本地文件系统
	StorageS3      StorageType = "s3"      // AWS S3
	StorageIPFS    StorageType = "ipfs"    // IPFS网络
)

// UniversalBlobReader 通用blob读取器实现
type UniversalBlobReader struct {
	config    *BlobReaderConfig
	client    *http.Client
	providers map[StorageType]StorageProvider
}

// BlobReaderConfig blob读取器配置
type BlobReaderConfig struct {
	DefaultTimeout    time.Duration // 默认超时时间
	MaxResponseSize   int64         // 最大响应大小
	UserAgent         string        // 用户代理
	DataPlaneEndpoint string        // ATProto数据平面端点
	VerifyChecksums   bool          // 是否验证校验和
	RetryAttempts     int           // 重试次数
}

// StorageProvider 存储提供者接口
type StorageProvider interface {
	GetData(ctx context.Context, options BlobOptions) (*BlobData, error)
	SupportsVerification() bool
}

// DefaultBlobReaderConfig 默认配置
func DefaultBlobReaderConfig() *BlobReaderConfig {
	return &BlobReaderConfig{
		DefaultTimeout:    30 * time.Second,
		MaxResponseSize:   100 << 20, // 100MB
		UserAgent:         "AvatarAI-BlobReader/1.0",
		DataPlaneEndpoint: "https://bsky.social",
		VerifyChecksums:   true,
		RetryAttempts:     3,
	}
}

// NewUniversalBlobReader 创建通用blob读取器
func NewUniversalBlobReader(config *BlobReaderConfig) *UniversalBlobReader {
	if config == nil {
		config = DefaultBlobReaderConfig()
	}

	client := &http.Client{
		Timeout: config.DefaultTimeout,
	}

	reader := &UniversalBlobReader{
		config:    config,
		client:    client,
		providers: make(map[StorageType]StorageProvider),
	}

	// 注册默认的存储提供者
	reader.RegisterProvider(StorageATProto, &ATProtoProvider{client: client, config: config})
	reader.RegisterProvider(StorageHTTP, &HTTPProvider{client: client, config: config})
	reader.RegisterProvider(StorageLocal, &LocalProvider{config: config})

	return reader
}

// RegisterProvider 注册存储提供者
func (r *UniversalBlobReader) RegisterProvider(storageType StorageType, provider StorageProvider) {
	r.providers[storageType] = provider
}

// StreamBlob 流式读取blob
func (r *UniversalBlobReader) StreamBlob(ctx context.Context, options BlobOptions, processor BlobProcessor) error {
	data, err := r.GetBlob(ctx, options)
	if err != nil {
		return err
	}

	return processor(data)
}

// GetBlob 获取blob数据
func (r *UniversalBlobReader) GetBlob(ctx context.Context, options BlobOptions) (*BlobData, error) {
	// 自动检测存储类型
	if options.StorageType == "" {
		options.StorageType = r.detectStorageType(options)
	}

	provider, exists := r.providers[options.StorageType]
	if !exists {
		return nil, xerrors.Errorf("不支持的存储类型: %s", options.StorageType)
	}

	// 获取数据
	data, err := provider.GetData(ctx, options)
	if err != nil {
		return nil, xerrors.Errorf("获取数据失败: %w", err)
	}

	// 验证数据
	if !options.SkipVerification && r.config.VerifyChecksums {
		if err := r.verifyData(data, options); err != nil {
			return nil, xerrors.Errorf("数据验证失败: %w", err)
		}
	}

	return data, nil
}

// VerifyBlob 验证blob
func (r *UniversalBlobReader) VerifyBlob(ctx context.Context, options BlobOptions) error {
	data, err := r.GetBlob(ctx, options)
	if err != nil {
		return err
	}

	return r.verifyData(data, options)
}

// detectStorageType 自动检测存储类型
func (r *UniversalBlobReader) detectStorageType(options BlobOptions) StorageType {
	identifier := options.Identifier

	// URL检测
	if parsedURL, err := url.Parse(identifier); err == nil && parsedURL.Scheme != "" {
		if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
			return StorageHTTP
		}
	}

	// CID检测
	if _, err := cid.Decode(identifier); err == nil {
		if options.DID != "" {
			return StorageATProto
		}
		return StorageIPFS
	}

	// 本地文件检测
	if identifier[0] == '/' || identifier[0] == '.' {
		return StorageLocal
	}

	// 默认为HTTP
	return StorageHTTP
}

// verifyData 验证数据完整性
func (r *UniversalBlobReader) verifyData(data *BlobData, options BlobOptions) error {
	// CID验证
	if options.CID != "" {
		return r.verifyCID(data.Data, options.CID)
	}

	// 大小验证
	if options.ExpectedSize > 0 && data.Size != options.ExpectedSize {
		return xerrors.Errorf("文件大小不匹配: 期望 %d, 实际 %d", options.ExpectedSize, data.Size)
	}

	return nil
}

// verifyCID 验证CID
func (r *UniversalBlobReader) verifyCID(data []byte, expectedCID string) error {
	expected, err := cid.Decode(expectedCID)
	if err != nil {
		return xerrors.Errorf("解析CID失败: %w", err)
	}

	// 使用SHA256计算哈希
	hasher, err := multihash.GetHasher(multihash.SHA2_256)
	if err != nil {
		return xerrors.Errorf("获取哈希器失败: %w", err)
	}

	hasher.Write(data)
	hash := hasher.Sum(nil)

	mh, err := multihash.Encode(hash, multihash.SHA2_256)
	if err != nil {
		return xerrors.Errorf("编码multihash失败: %w", err)
	}

	actual := cid.NewCidV1(cid.Raw, mh)

	if !actual.Equals(expected) {
		return xerrors.Errorf("CID不匹配: 期望 %s, 实际 %s", expected.String(), actual.String())
	}

	return nil
}

// ===========================================
// 存储提供者实现
// ===========================================

// ATProtoProvider ATProto存储提供者
type ATProtoProvider struct {
	client *http.Client
	config *BlobReaderConfig
}

func (p *ATProtoProvider) GetData(ctx context.Context, options BlobOptions) (*BlobData, error) {
	blobURL := fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s",
		p.config.DataPlaneEndpoint, options.DID, options.CID)

	req, err := http.NewRequestWithContext(ctx, "GET", blobURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", p.config.UserAgent)
	if options.AcceptEncoding != "" {
		req.Header.Set("Accept-Encoding", options.AcceptEncoding)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, p.config.MaxResponseSize)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &BlobData{
		Data:        data,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        int64(len(data)),
		Source:      blobURL,
		Metadata: map[string]string{
			"did": options.DID,
			"cid": options.CID,
		},
	}, nil
}

func (p *ATProtoProvider) SupportsVerification() bool {
	return true
}

// HTTPProvider HTTP存储提供者
type HTTPProvider struct {
	client *http.Client
	config *BlobReaderConfig
}

func (p *HTTPProvider) GetData(ctx context.Context, options BlobOptions) (*BlobData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", options.Identifier, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", p.config.UserAgent)
	for k, v := range options.Headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, p.config.MaxResponseSize)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &BlobData{
		Data:        data,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        int64(len(data)),
		Source:      options.Identifier,
	}, nil
}

func (p *HTTPProvider) SupportsVerification() bool {
	return false
}

// LocalProvider 本地文件存储提供者
type LocalProvider struct {
	config *BlobReaderConfig
}

func (p *LocalProvider) GetData(ctx context.Context, options BlobOptions) (*BlobData, error) {
	data, err := os.ReadFile(options.Identifier)
	if err != nil {
		return nil, xerrors.Errorf("读取本地文件失败: %w", err)
	}

	return &BlobData{
		Data:        data,
		ContentType: "application/octet-stream", // 默认类型
		Size:        int64(len(data)),
		Source:      options.Identifier,
	}, nil
}

func (p *LocalProvider) SupportsVerification() bool {
	return true
}
