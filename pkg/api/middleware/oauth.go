package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
	"github.com/zhongshangwu/avatarai-social/types"
	"gorm.io/gorm"
)

// WrapHandler 将 types.APIContext 处理函数适配为 echo.HandlerFunc
func WrapHandler(handler func(*types.APIContext) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 将 echo.Context 转换为 types.APIContext
		apiCtx := &types.APIContext{
			Context: c,
		}

		// 如果是 AvatarAIContext，提取认证信息
		if avatarCtx, ok := c.(*utils.AvatarAIContext); ok {
			apiCtx.IsAuthenticated = avatarCtx.Avatar != nil
			if avatarCtx.Avatar != nil {
				// 构建头像和横幅 URL
				avatarURL := ""
				if avatarCtx.Avatar.AvatarCID != "" {
					avatarURL = fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg",
						avatarCtx.Avatar.Did,
						avatarCtx.Avatar.AvatarCID)
				}

				apiCtx.User = &types.User{
					Did:         avatarCtx.Avatar.Did,
					Handle:      avatarCtx.Avatar.Handle,
					DisplayName: avatarCtx.Avatar.DisplayName,
					Description: avatarCtx.Avatar.Description,
					AvatarCID:   avatarCtx.Avatar.AvatarCID,
					AvatarURL:   avatarURL,
					BannerCID:   "", // repositories.Avatar 没有 BannerCID 字段
					BannerURL:   "", // repositories.Avatar 没有 BannerURL 字段
					IsAster:     avatarCtx.Avatar.IsAster,
					Creator:     avatarCtx.Avatar.CreatorDid,
					LastLoginAt: avatarCtx.Avatar.LastLoginAt.Unix(),
					UpdatedAt:   avatarCtx.Avatar.UpdatedAt.Unix(),
					CreatedAt:   avatarCtx.Avatar.CreatedAt.Unix(),
				}
			}

			// 转换 OAuthSession
			if avatarCtx.OauthSession != nil {
				apiCtx.OauthSession = &types.OAuthSession{
					ID:                  fmt.Sprintf("%d", avatarCtx.OauthSession.ID),
					Did:                 avatarCtx.OauthSession.Did,
					Handle:              avatarCtx.OauthSession.Handle,
					PdsUrl:              avatarCtx.OauthSession.PdsUrl,
					AuthserverIss:       avatarCtx.OauthSession.AuthserverIss,
					AccessToken:         avatarCtx.OauthSession.AccessToken,
					RefreshToken:        avatarCtx.OauthSession.RefreshToken,
					DpopAuthserverNonce: avatarCtx.OauthSession.DpopAuthserverNonce,
					DpopPdsNonce:        avatarCtx.OauthSession.DpopPdsNonce,
					DpopPrivateJwk:      avatarCtx.OauthSession.DpopPrivateJwk,
					ExpiresIn:           avatarCtx.OauthSession.ExpiresIn,
					CreatedAt:           avatarCtx.OauthSession.CreatedAt.Unix(),
					Provider:            types.OAuthProviderTypeBsky,
					ReturnURI:           avatarCtx.OauthSession.ReturnURI,
				}
			}
		}

		return handler(apiCtx)
	}
}

func NewRequireAuth(metaStore *repositories.MetaStore, config *config.SocialConfig) echo.MiddlewareFunc {
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

			session, err := metaStore.UserRepo.GetSessionByID(claims.ID)
			if err != nil {
				log.Errorf("获取Session失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			oauthSession, err := metaStore.OAuthRepo.GetOAuthSessionByID(session.OAuthSessionID)
			if err != nil {
				log.Errorf("获取OAuthSession失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			avatar, err := metaStore.UserRepo.GetAvatarByDID(session.AvatarDid)
			if err != nil {
				log.Errorf("获取Avatar失败: %v", err)
				return cc.Redirect(http.StatusFound, "/api/oauth/login")
			}

			cc.Session = session
			cc.OauthSession = oauthSession
			cc.Avatar = avatar

			// 检查 OAuth 会话是否过期
			if oauthSession.ExpiresIn > 0 {
				expiryTime := oauthSession.CreatedAt.Add(time.Duration(oauthSession.ExpiresIn) * time.Second)
				if time.Now().After(expiryTime) {
					log.Warnf("OAuth会话已过期，需要刷新令牌")
					// 这里可以尝试自动刷新令牌，或者重定向到登录页面
					return cc.Redirect(http.StatusFound, "/api/oauth/login")
				}
			}

			return next(cc)
		}
	}
}

func IsSessionExpired(session *repositories.OAuthSession) bool {
	log.Infof("session.CreatedAt: %s", session.CreatedAt.Unix())
	log.Infof("session.ExpiresIn: %d", session.ExpiresIn)
	log.Infof("time.Now().Unix(): %d", time.Now().Unix())
	log.Infof("expired: %t", time.Now().Unix() > session.CreatedAt.Unix()+int64(session.ExpiresIn))
	return time.Now().Unix() > session.CreatedAt.Unix()+int64(session.ExpiresIn)
}

func RefreshSession(c echo.Context, db *gorm.DB, session *repositories.OAuthSession, config *config.SocialConfig) error {
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

	// 这里需要通过 MetaStore 来更新，但我们只有 db，所以保持原来的方式
	// 或者我们需要传递 MetaStore 而不是 db
	if err := db.Model(&repositories.OAuthSession{}).
		Where("did = ?", session.Did).
		Updates(map[string]interface{}{
			"access_token":          session.AccessToken,
			"refresh_token":         session.RefreshToken,
			"dpop_authserver_nonce": session.DpopAuthserverNonce,
		}).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "更新会话失败",
		})
	}
	return nil
}
