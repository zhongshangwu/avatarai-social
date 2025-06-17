package blobs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"image"
	_ "image/gif" // 支持GIF解码
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"golang.org/x/xerrors"
)

type ImageViewer struct {
	config  *ImageViewerConfig
	cache   BlobCache
	client  *http.Client
	builder *ImageUriBuilder
}

type ImageViewerConfig struct {
	CacheLocation         string        // 缓存目录
	CDNUrl                string        // CDN URL，如果设置了则不提供图片服务
	MaxResponseSize       int64         // 最大响应大小
	HeadersTimeout        time.Duration // 头部超时
	BodyTimeout           time.Duration // 正文超时
	MaxRetries            int           // 最大重试次数
	DisableSSRFProtection bool          // 是否禁用SSRF保护
	ProxyPreferCompressed bool          // 是否优先压缩
	RateLimitBypassKey    string        // 速率限制绕过密钥
	RateLimitBypassHost   string        // 速率限制绕过主机
	UserAgent             string        // 用户代理
}

func DefaultImageViewerConfig() *ImageViewerConfig {
	return &ImageViewerConfig{
		CacheLocation:   "/tmp/avatarai-images-cache",
		MaxResponseSize: 50 << 20, // 50MB
		HeadersTimeout:  30 * time.Second,
		BodyTimeout:     60 * time.Second,
		MaxRetries:      3,
		UserAgent:       "AvatarAI-Social/1.0",
	}
}

type StreamBlobOptions struct {
	DID            string
	CID            string
	AcceptEncoding string
	Context        context.Context
}

func NewImageViewer(config *ImageViewerConfig) (*ImageViewer, error) {
	if config == nil {
		config = DefaultImageViewerConfig()
	}

	cache, err := NewDiskCache(config.CacheLocation)
	if err != nil {
		return nil, xerrors.Errorf("创建缓存失败: %w", err)
	}

	client := &http.Client{
		Timeout: config.HeadersTimeout + config.BodyTimeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: config.HeadersTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &ImageViewer{
		config:  config,
		cache:   cache,
		client:  client,
		builder: NewImageUriBuilder(""),
	}, nil
}

func (v *ImageViewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware := v.CreateMiddleware("/")
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})).ServeHTTP(w, r)
}

func (v *ImageViewer) StreamBlob(ctx context.Context, options StreamBlobOptions, factory func([]byte, BlobInfo) ([]byte, error)) error {
	blobURL, err := v.getBlobURL(ctx, options.DID, options.CID)
	if err != nil {
		return xerrors.Errorf("获取blob URL失败: %w", err)
	}

	data, contentType, err := v.fetchBlob(ctx, blobURL)
	if err != nil {
		return err
	}

	cidObj, err := cid.Decode(options.CID)
	if err != nil {
		return xerrors.Errorf("解析CID失败: %w", err)
	}

	if err := v.verifyCID(data, options.CID); err != nil {
		return err
	}

	parsedURL, _ := url.Parse(blobURL)
	blobInfo := BlobInfo{
		URL:    parsedURL,
		DID:    options.DID,
		CID:    cidObj,
		Size:   int64(len(data)),
		Format: contentType,
	}

	_, err = factory(data, blobInfo)
	return err
}

func (v *ImageViewer) CreateMiddleware(prefix string) func(http.Handler) http.Handler {
	if !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") {
		panic("前缀必须以/开始和结束")
	}

	// 如果有CDN，不提供图片服务
	if v.config.CDNUrl != "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logrus.Infof("ImageViewer: %s", r.URL.Path)
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path[len(prefix)-1:]
			if !strings.HasPrefix(path, "/") || path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			v.handleImageRequest(w, r, path)
		})
	}
}

func (v *ImageViewer) handleImageRequest(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	logrus.Infof("ImageViewer: %s...", path)

	// 解析路径
	options, blobLoc, preset, err := GetOptions(path)
	if err != nil {
		http.Error(w, "无效的路径", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("%s::%s::%s", blobLoc.DID, blobLoc.CID, preset)

	// 检查缓存
	if cached, err := v.cache.Get(cacheKey); err == nil {
		v.serveCachedImage(w, cached)
		return
	}

	// 获取并处理图片
	if err := v.processAndServeImage(ctx, w, options, blobLoc, cacheKey); err != nil {
		v.handleError(w, err)
	}
}

func (v *ImageViewer) serveCachedImage(w http.ResponseWriter, cached *CachedBlob) {
	w.Header().Set("X-Cache", "hit")
	w.Header().Set("Content-Type", cached.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(cached.Size, 10))
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1年
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "0")

	w.WriteHeader(http.StatusOK)
	w.Write(cached.Data)
}

func (v *ImageViewer) processAndServeImage(ctx context.Context, w http.ResponseWriter, options *Options, blobLoc *BlobLocation, cacheKey string) error {
	blobURL, err := v.getBlobURL(ctx, blobLoc.DID, blobLoc.CID)
	if err != nil {
		return xerrors.Errorf("获取blob URL失败: %w", err)
	}

	logrus.Infof("fetchBlob: %s", blobURL)
	data, contentType, err := v.fetchBlob(ctx, blobURL)
	if err != nil {
		logrus.Errorf("fetchBlob error: %v", err)
		return xerrors.Errorf("获取blob失败: %w", err)
	}
	logrus.Infof("fetchBlob success, length: %d, contentType: %s", len(data), contentType)

	// 验证是否为图片
	if !v.isImageMime(contentType) {
		return xerrors.New("不是图片类型")
	}

	// 验证CID
	if err := v.verifyCID(data, blobLoc.CID); err != nil {
		return xerrors.Errorf("CID验证失败: %w", err)
	}

	// 处理图片
	processedData, processedType, err := v.processImage(data, options)
	if err != nil {
		return xerrors.Errorf("图片处理失败: %w", err)
	}

	// 异步缓存
	go func() {
		if err := v.cache.Put(cacheKey, processedData, processedType); err != nil {
			// 记录错误，但不影响响应
			logrus.Errorf("缓存图片失败: %v\n", err)
		}
	}()

	// 响应
	w.Header().Set("X-Cache", "miss")
	w.Header().Set("Content-Type", processedType)
	w.Header().Set("Content-Length", strconv.Itoa(len(processedData)))
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1年
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "0")

	w.WriteHeader(http.StatusOK)
	w.Write(processedData)

	return nil
}

func (v *ImageViewer) fetchBlob(ctx context.Context, blobURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blobURL, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("User-Agent", v.config.UserAgent)

	// 添加速率限制绕过头部
	if v.config.RateLimitBypassKey != "" && v.config.RateLimitBypassHost != "" {
		parsedURL, _ := url.Parse(blobURL)
		if parsedURL != nil &&
			(strings.HasSuffix(parsedURL.Hostname(), v.config.RateLimitBypassHost) ||
				parsedURL.Hostname() == v.config.RateLimitBypassHost) {
			req.Header.Set("X-Ratelimit-Bypass", v.config.RateLimitBypassKey)
		}
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, "", xerrors.Errorf("Blob not found: %d", resp.StatusCode)
		}
		return nil, "", xerrors.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 限制响应大小
	reader := io.LimitReader(resp.Body, v.config.MaxResponseSize)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return data, contentType, nil
}

func (v *ImageViewer) isImageMime(contentType string) bool {
	if contentType == "" || contentType == "application/octet-stream" {
		return false // 不确定的情况下返回false
	}
	return strings.HasPrefix(contentType, "image/")
}

func (v *ImageViewer) verifyCID(data []byte, expectedCID string) error {
	// 解析预期的CID
	expected, err := cid.Decode(expectedCID)
	if err != nil {
		return xerrors.Errorf("解析CID失败: %w", err)
	}

	// 计算数据的哈希
	hash := sha256.Sum256(data)

	// 创建multihash
	mh, err := multihash.EncodeName(hash[:], "sha2-256")
	if err != nil {
		return xerrors.Errorf("创建multihash失败: %w", err)
	}

	// 创建CID
	actual := cid.NewCidV1(cid.Raw, mh)

	// 比较CID
	if !actual.Equals(expected) {
		return xerrors.Errorf("CID不匹配: 期望 %s, 实际 %s", expected.String(), actual.String())
	}

	return nil
}

func (v *ImageViewer) processImage(data []byte, options *Options) ([]byte, string, error) {
	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", xerrors.Errorf("解码图片失败: %w", err)
	}

	// 简化处理：对于现在的实现，主要是格式转换
	// 如果需要缩放等高级功能，需要添加图片处理库

	var buf bytes.Buffer
	var contentType string

	// 根据选项决定输出格式
	outputFormat := options.Format
	if outputFormat == "" {
		outputFormat = format
	}

	switch outputFormat {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		contentType = "image/jpeg"
	case "png":
		err = png.Encode(&buf, img)
		contentType = "image/png"
	default:
		// 默认使用JPEG
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		contentType = "image/jpeg"
	}

	if err != nil {
		return nil, "", xerrors.Errorf("编码图片失败: %w", err)
	}

	return buf.Bytes(), contentType, nil
}

func (v *ImageViewer) handleError(w http.ResponseWriter, err error) {
	errStr := err.Error()
	if strings.Contains(errStr, "无效的路径") {
		http.Error(w, "Bad Path", http.StatusBadRequest)
	} else if strings.Contains(errStr, "CID") || strings.Contains(errStr, "Blob not found") {
		http.Error(w, "Blob not found", http.StatusNotFound)
	} else if strings.Contains(errStr, "不是图片") {
		http.Error(w, "Not an image", http.StatusBadRequest)
	} else {
		http.Error(w, "Upstream Error", http.StatusBadGateway)
	}
}

func (v *ImageViewer) getBlobURL(ctx context.Context, did string, cid string) (string, error) {
	pdsURL, err := v.resolvePDS(ctx, did)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s", pdsURL, did, cid), nil
}

func (v *ImageViewer) resolvePDS(ctx context.Context, did string) (string, error) {
	ident, err := atproto.ResolveIdentity(ctx, did)
	if err != nil {
		return "", err
	}
	return atproto.PDSEndpoint(ident), nil
}
