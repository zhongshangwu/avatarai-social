package types

type User struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	PdsUrl      string `json:"pdsUrl"`
	DisplayName string `json:"displayName"`
	AvatarCID   string `json:"avatarId"`
	AvatarURL   string `json:"avatarUrl"`
	BannerCID   string `json:"bannerId"`
	BannerURL   string `json:"bannerUrl"`
	Description string `json:"description"`
	IsAster     bool   `json:"isAster"`
	Creator     string `json:"creator"` // 对于 aster 来说, creator 是真实的人, 对于 avatar 来说, creator 是 avatar 自己
	LastLoginAt int64  `json:"lastLoginAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	CreatedAt   int64  `json:"createdAt"`
}

type Session struct {
	ID             string            `json:"id"`
	UserDid        string            `json:"userDid"`
	AccessToken    string            `json:"accessToken"`
	RefreshToken   string            `json:"refreshToken"`
	OAuthSessionID string            `json:"oauthSessionId"`
	OAuthProvider  OAuthProviderType `json:"oauthProvider"`
	ExpiredAt      int64             `json:"expiredAt"`
	CreatedAt      int64             `json:"createdAt"`
	UpdatedAt      int64             `json:"updatedAt"`
}

type OAuthSession struct {
	ID                  string            `json:"id"`
	Did                 string            `json:"did"`
	Handle              string            `json:"handle"`
	PdsUrl              string            `json:"pdsUrl"`
	AuthserverIss       string            `json:"authserverIss"`
	AccessToken         string            `json:"accessToken"`
	RefreshToken        string            `json:"refreshToken"`
	DpopAuthserverNonce string            `json:"dpopAuthserverNonce"`
	DpopPdsNonce        string            `json:"dpopPdsNonce"`
	DpopPrivateJwk      string            `json:"dpopPrivateJwk"`
	ExpiresIn           int64             `json:"expiresIn"`
	CreatedAt           int64             `json:"createdAt"`
	Provider            OAuthProviderType `json:"provider"`
	ReturnURI           string            `json:"returnURI"`
}

type OAuthProviderType string

const (
	OAuthProviderTypeBsky OAuthProviderType = "bsky"
)
