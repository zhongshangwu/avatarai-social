package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
	"gorm.io/gorm"
)

func NewRequireAuth(metaStore *database.MetaStore, config *config.SocialConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &utils.AvatarAIContext{Context: c}
			var token string
			jwtCookie, err := cc.Cookie("avatarai_token")
			if err != nil || jwtCookie.Value == "" {
				// 尝试从 Authorization header 中获取
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" {
					parts := strings.Split(authHeader, " ")
					if len(parts) == 2 && parts[0] == "Bearer" {
						token = parts[1]
					}
				}
				if token == "" {
					return cc.Redirect(http.StatusFound, "/api/oauth/login")
				}
			} else {
				token = jwtCookie.Value
			}

			claims, err := utils.ValidateAccessToken(config, token)
			if err != nil {
				log.Errorf("验证token失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			session, err := database.GetSessionByID(metaStore.DB, claims.ID)
			if err != nil {
				log.Errorf("获取Session失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			oauthSession, err := database.GetOauthSessionByID(metaStore.DB, session.OAuthSessionID)
			if err != nil {
				log.Errorf("获取OAuthSession失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			avatar, err := database.GetAvatarByDID(metaStore.DB, session.AvatarDid)
			if err != nil {
				log.Errorf("获取Avatar失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			cc.Session = session
			cc.OauthSession = oauthSession
			cc.Avatar = avatar
			return next(cc)
		}
	}
}

func IsSessionExpired(session *database.OAuthSession) bool {
	log.Infof("session.CreatedAt: %s", session.CreatedAt.Unix())
	log.Infof("session.ExpiresIn: %d", session.ExpiresIn)
	log.Infof("time.Now().Unix(): %d", time.Now().Unix())
	log.Infof("expired: %t", time.Now().Unix() > session.CreatedAt.Unix()+int64(session.ExpiresIn))
	return time.Now().Unix() > session.CreatedAt.Unix()+int64(session.ExpiresIn)
}

func RefreshSession(c echo.Context, db *gorm.DB, session *database.OAuthSession, config *config.SocialConfig) error {
	appURL := utils.GetAPPURL(c)
	tokens, dpopAuthserverNonce, err := atproto.RefreshTokenRequest(
		session,
		atproto.BuildClientID(appURL, session.Platform),
		atproto.BuildRedirectURL(appURL, session.Platform),
		config.ATP.ClientSecretJWK(),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "刷新令牌失败: " + err.Error(),
		})
	}

	// 将更新的令牌(和 DPoP nonce)保存到数据库
	session.AccessToken = tokens["access_token"].(string)
	session.RefreshToken = tokens["refresh_token"].(string)
	session.DpopAuthserverNonce = dpopAuthserverNonce

	if err := database.UpdateOAuthSession(db, session); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "更新会话失败",
		})
	}
	return nil
}
