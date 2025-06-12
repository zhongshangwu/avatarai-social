package blobs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
)

type BlobInfo struct {
	URL    *url.URL
	DID    string
	CID    cid.Cid
	Size   int64
	Format string
}

type BlobCache interface {
	Get(key string) (*CachedBlob, error)
	Put(key string, data []byte, contentType string) error
	Clear(key string) error
	ClearAll() error
}

type CachedBlob struct {
	Data        []byte
	ContentType string
	Size        int64
	CachedAt    time.Time
}

type DiskCache struct {
	basePath string
	mu       sync.RWMutex
}

func NewDiskCache(basePath string) (*DiskCache, error) {
	if !filepath.IsAbs(basePath) {
		return nil, xerrors.New("必须提供绝对路径")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, xerrors.Errorf("创建缓存目录失败: %w", err)
	}

	return &DiskCache{basePath: basePath}, nil
}

func (c *DiskCache) Get(key string) (*CachedBlob, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	safeKey := c.safeKey(key)
	filePath := filepath.Join(c.basePath, safeKey)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, xerrors.New("缓存不存在")
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, xerrors.New("缓存为空")
	}

	// 简单的元数据存储格式：contentType\n\ndata
	parts := bytes.SplitN(data, []byte("\n\n"), 2)
	if len(parts) != 2 {
		return nil, xerrors.New("缓存格式错误")
	}

	return &CachedBlob{
		Data:        parts[1],
		ContentType: string(parts[0]),
		Size:        int64(len(parts[1])),
		CachedAt:    time.Now(),
	}, nil
}

func (c *DiskCache) Put(key string, data []byte, contentType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	safeKey := c.safeKey(key)
	filePath := filepath.Join(c.basePath, safeKey)

	// 创建包含元数据的内容
	content := append([]byte(contentType+"\n\n"), data...)

	return os.WriteFile(filePath, content, 0644)
}

func (c *DiskCache) Clear(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	safeKey := c.safeKey(key)
	filePath := filepath.Join(c.basePath, safeKey)

	return os.Remove(filePath)
}

func (c *DiskCache) ClearAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return os.RemoveAll(c.basePath)
}

func (c *DiskCache) safeKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
