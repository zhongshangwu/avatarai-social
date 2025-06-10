package atproto

import (
	"fmt"
	"strings"
)

var AllowedPlatforms = []string{"ios", "android", "web"}

type OAuthClientMetadata struct {
	RedirectURIs                          []string `json:"redirect_uris"`
	ResponseTypes                         []string `json:"response_types,omitempty"`
	GrantTypes                            []string `json:"grant_types,omitempty"`
	Scope                                 string   `json:"scope,omitempty"`
	TokenEndpointAuthMethod               string   `json:"token_endpoint_auth_method,omitempty"`
	TokenEndpointAuthSigningAlg           string   `json:"token_endpoint_auth_signing_alg,omitempty"`
	UserinfoSignedResponseAlg             string   `json:"userinfo_signed_response_alg,omitempty"`
	UserinfoEncryptedResponseAlg          string   `json:"userinfo_encrypted_response_alg,omitempty"`
	JwksURI                               string   `json:"jwks_uri,omitempty"`
	ApplicationType                       string   `json:"application_type,omitempty"` // "web" or "native"
	SubjectType                           string   `json:"subject_type,omitempty"`     // "public" or "pairwise"
	RequestObjectSigningAlg               string   `json:"request_object_signing_alg,omitempty"`
	IDTokenSignedResponseAlg              string   `json:"id_token_signed_response_alg,omitempty"`
	AuthorizationSignedResponseAlg        string   `json:"authorization_signed_response_alg,omitempty"`
	AuthorizationEncryptedResponseEnc     string   `json:"authorization_encrypted_response_enc,omitempty"`
	AuthorizationEncryptedResponseAlg     string   `json:"authorization_encrypted_response_alg,omitempty"`
	ClientID                              string   `json:"client_id,omitempty"`
	ClientName                            string   `json:"client_name,omitempty"`
	ClientURI                             string   `json:"client_uri,omitempty"`
	PolicyURI                             string   `json:"policy_uri,omitempty"`
	TosURI                                string   `json:"tos_uri,omitempty"`
	LogoURI                               string   `json:"logo_uri,omitempty"`
	DefaultMaxAge                         int      `json:"default_max_age,omitempty"`
	RequireAuthTime                       *bool    `json:"require_auth_time,omitempty"`
	Contacts                              []string `json:"contacts,omitempty"`
	TLSClientCertificateBoundAccessTokens *bool    `json:"tls_client_certificate_bound_access_tokens,omitempty"`
	DPoPBoundAccessTokens                 *bool    `json:"dpop_bound_access_tokens,omitempty"`
	AuthorizationDetailsTypes             []string `json:"authorization_details_types,omitempty"`
	// Jwks                                  *JWKSet  `json:"jwks,omitempty"`             // You'll need to define JWKSet type
}

type ClientMetadataOptions struct {
	ClientName              string // 客户端名称，默认为 "AvatarAI Social"
	TokenEndpointAuthMethod string // 认证方法，默认为 "private_key_jwt"
	JwksURI                 string // JWKS URI，可选
	UsePrivateKeyJWT        bool   // 是否使用私钥JWT认证，默认为 true
}

func GetClientMetadata(host string, platform string, appBundleId string) *OAuthClientMetadata {
	return GetClientMetadataWithOptions(host, platform, appBundleId, nil)
}

func GetClientMetadataWithOptions(host string, platform string, appBundleId string, options *ClientMetadataOptions) *OAuthClientMetadata {
	// 设置默认选项
	if options == nil {
		options = &ClientMetadataOptions{}
	}
	if options.ClientName == "" {
		options.ClientName = "AvatarAI Social"
	}
	if options.TokenEndpointAuthMethod == "" {
		if options.UsePrivateKeyJWT {
			options.TokenEndpointAuthMethod = "private_key_jwt"
		} else {
			options.TokenEndpointAuthMethod = "none"
		}
	}

	meta := &OAuthClientMetadata{
		ClientID:                BuildClientID(host, platform),
		ClientURI:               fmt.Sprintf("https://%s", host),
		Scope:                   "atproto transition:generic",
		TokenEndpointAuthMethod: options.TokenEndpointAuthMethod,
		ClientName:              options.ClientName,
		ResponseTypes:           []string{"code"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		DPoPBoundAccessTokens:   boolPtr(true),
	}

	// 如果使用私钥JWT认证，设置相关参数
	if options.TokenEndpointAuthMethod == "private_key_jwt" {
		meta.TokenEndpointAuthSigningAlg = "ES256"
		if options.JwksURI != "" {
			meta.JwksURI = options.JwksURI
		} else {
			meta.JwksURI = fmt.Sprintf("https://%s/api/oauth/jwks.json", host)
		}
	}

	// 根据平台设置重定向URI和应用类型
	if platform == "web" {
		meta.RedirectURIs = []string{fmt.Sprintf("https://%s/api/oauth/callback", host)}
		meta.ApplicationType = "web"
	} else {
		meta.RedirectURIs = []string{fmt.Sprintf("https://%s/api/app-return/%s", host, appBundleId)}
		meta.ApplicationType = "native"
	}
	return meta
}

func BuildClientID(host string, platform string) string {
	// 兼容 localhost 开发环境
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		return fmt.Sprintf("http://localhost/?scope=atproto transition:generic&redirect_uri=%sapi/oauth/callback", host)
	}
	return fmt.Sprintf("%sapi/oauth/%s/client-metadata.json", host, platform)
}

func BuildRedirectURL(host string, platform string) string {
	return fmt.Sprintf("%sapi/oauth/callback", host)
}

func BuildCallbackRedirectURI(host string, platform string) string {
	if platform == "web" {
		return fmt.Sprintf("https://%s/login", host)
	}
	return fmt.Sprintf("https://%s/app-callback", host)
}

func boolPtr(b bool) *bool {
	return &b
}
