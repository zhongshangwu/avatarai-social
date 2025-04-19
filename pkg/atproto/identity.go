package atproto

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/carlmjohnson/versioninfo"
)

// 正则表达式常量，用于验证 handle 和 DID
const (
	HandleRegex = `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`

	DIDRegex = `^did:[a-z]+:[a-zA-Z0-9._:%-]*[a-zA-Z0-9._-]$`

	DefaultPLCURL = "https://plc.avatar.ai"
)

// 预编译正则表达式以提高性能
var (
	handleRegex = regexp.MustCompile(HandleRegex)
	didRegex    = regexp.MustCompile(DIDRegex)
)

// IsValidHandle 检查给定的字符串是否为有效的 ATProto handle
// handle 格式应该类似域名，如 "user.bsky.social"
func IsValidHandle(handle string) bool {
	return handleRegex.MatchString(handle)
}

// IsValidDID 检查给定的字符串是否为有效的分布式标识符(DID)
// DID 格式应该如 "did:method:specific-id-value"
func IsValidDID(did string) bool {
	return didRegex.MatchString(did)
}

func ResolveIdentity(ctx context.Context, arg string) (*identity.Identity, error) {
	id, err := syntax.ParseAtIdentifier(arg)
	if err != nil {
		return nil, err
	}

	dir := DefaultDirectory()
	return dir.Lookup(ctx, *id)
}

func PDSEndpoint(ident *identity.Identity) string {
	var svc *identity.Service
	for _, s := range ident.Services {
		if s.Type == "AtprotoPersonalDataServer" {
			svc = &s
			break
		}
	}

	if svc == nil {
		return ""
	}

	return svc.URL
}

func DefaultDirectory() identity.Directory {
	base := identity.BaseDirectory{
		PLCURL: DefaultPLCURL,
		HTTPClient: http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				// would want this around 100ms for services doing lots of handle resolution. Impacts PLC connections as well, but not too bad.
				IdleConnTimeout: time.Millisecond * 1000,
				MaxIdleConns:    100,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		Resolver: net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: time.Second * 3}
				return d.DialContext(ctx, network, address)
			},
		},
		TryAuthoritativeDNS: true,
		// primary Bluesky PDS instance only supports HTTP resolution method
		SkipDNSDomainSuffixes: []string{".bsky.social"},
		UserAgent:             "indigo-identity/" + versioninfo.Short(),
	}
	cached := identity.NewCacheDirectory(&base, 250_000, time.Hour*24, time.Minute*2, time.Minute*5)
	return &cached
}
